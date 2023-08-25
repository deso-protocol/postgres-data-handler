package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PostAssociationEntry struct {
	AssociationID    string `bun:",nullzero"`
	TransactorPKID   string `bun:",nullzero"`
	PostHash         string `bun:",nullzero"`
	AppPKID          string `bun:",nullzero"`
	AssociationType  string `pg:",use_zero"`
	AssociationValue string `pg:",use_zero"`
	BlockHeight      uint32 `bun:",nullzero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGPostAssociationEntry struct {
	bun.BaseModel `bun:"table:post_association_entry"`
	PostAssociationEntry
}

type PGPostAssociationEntryUtxoOps struct {
	bun.BaseModel `bun:"table:post_association_entry_utxo_ops"`
	PostAssociationEntry
	UtxoOperation
}

// Convert the PostAssociation DeSo encoder to the PG struct used by bun.
func PostAssociationEncoderToPGStruct(postAssociationEntry *lib.PostAssociationEntry, keyBytes []byte, params *lib.DeSoParams) PostAssociationEntry {
	pgEntry := PostAssociationEntry{
		AssociationType:  string(postAssociationEntry.AssociationType[:]),
		AssociationValue: string(postAssociationEntry.AssociationValue[:]),
		ExtraData:        consumer.ExtraDataBytesToString(postAssociationEntry.ExtraData),
		BadgerKey:        keyBytes,
	}
	if postAssociationEntry.AssociationID != nil {
		pgEntry.AssociationID = hex.EncodeToString(postAssociationEntry.AssociationID[:])
	}
	if postAssociationEntry.TransactorPKID != nil {
		pgEntry.TransactorPKID = consumer.PublicKeyBytesToBase58Check(postAssociationEntry.TransactorPKID[:], params)
	}

	if postAssociationEntry.PostHash != nil {
		pgEntry.PostHash = hex.EncodeToString(postAssociationEntry.PostHash[:])
	}

	if postAssociationEntry.AppPKID != nil {
		pgEntry.AppPKID = consumer.PublicKeyBytesToBase58Check(postAssociationEntry.AppPKID[:], params)
	}
	return pgEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func PostAssociationBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeletePostAssociationEntry(entries, db, operationType)
	} else {
		err = bulkInsertPostAssociationEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertPostAssociationEntry inserts a batch of post_association entries into the database.
func bulkInsertPostAssociationEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGPostAssociationEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGPostAssociationEntry{PostAssociationEntry: PostAssociationEncoderToPGStruct(entry.Encoder.(*lib.PostAssociationEntry), entry.KeyBytes, params)}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertPostAssociationEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of post_association entries from the database.
func bulkDeletePostAssociationEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGPostAssociationEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeletePostAssociationEntry: Error deleting entries")
	}

	return nil
}
