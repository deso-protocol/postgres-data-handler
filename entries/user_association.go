package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type UserAssociationEntry struct {
	AssociationID    string `bun:",nullzero"`
	TransactorPKID   string `bun:",nullzero"`
	TargetUserPKID   string `bun:",nullzero"`
	AppPKID          string `bun:",nullzero"`
	AssociationType  string `pg:",use_zero"`
	AssociationValue string `pg:",use_zero"`
	BlockHeight      uint32 `bun:",nullzero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGUserAssociationEntry struct {
	bun.BaseModel `bun:"table:user_association_entry"`
	UserAssociationEntry
}

type PGUserAssociationEntryUtxoOps struct {
	bun.BaseModel `bun:"table:user_association_entry_utxo_ops"`
	UserAssociationEntry
	UtxoOperation
}

// Convert the UserAssociation DeSo encoder to the PG struct used by bun.
func UserAssociationEncoderToPGStruct(userAssociationEntry *lib.UserAssociationEntry, keyBytes []byte, params *lib.DeSoParams) UserAssociationEntry {
	pgEntry := UserAssociationEntry{
		AssociationType:  string(userAssociationEntry.AssociationType[:]),
		AssociationValue: string(userAssociationEntry.AssociationValue[:]),
		ExtraData:        consumer.ExtraDataBytesToString(userAssociationEntry.ExtraData),
		BadgerKey:        keyBytes,
	}
	if userAssociationEntry.AssociationID != nil {
		pgEntry.AssociationID = hex.EncodeToString(userAssociationEntry.AssociationID[:])
	}
	if userAssociationEntry.TransactorPKID != nil {
		pgEntry.TransactorPKID = consumer.PublicKeyBytesToBase58Check(userAssociationEntry.TransactorPKID[:], params)
	}

	if userAssociationEntry.TargetUserPKID != nil {
		pgEntry.TargetUserPKID = consumer.PublicKeyBytesToBase58Check(userAssociationEntry.TargetUserPKID[:], params)
	}

	if userAssociationEntry.AppPKID != nil {
		pgEntry.AppPKID = consumer.PublicKeyBytesToBase58Check(userAssociationEntry.AppPKID[:], params)
	}
	return pgEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate methods
// based on the operation type and executes it.
func UserAssociationBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteUserAssociationEntry(entries, db, operationType)
	} else {
		err = bulkInsertUserAssociationEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertUserAssociationEntry inserts a batch of user_association entries into the database.
func bulkInsertUserAssociationEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGUserAssociationEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGUserAssociationEntry{UserAssociationEntry: UserAssociationEncoderToPGStruct(entry.Encoder.(*lib.UserAssociationEntry), entry.KeyBytes, params)}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertUserAssociationEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of user_association entries from the database.
func bulkDeleteUserAssociationEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGUserAssociationEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteUserAssociationEntry: Error deleting entries")
	}

	return nil
}
