package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type AccessGroupEntry struct {
	AccessGroupOwnerPublicKey string `bun:",nullzero"`
	AccessGroupKeyName        string `pg:",use_zero"`
	AccessGroupPublicKey      string `bun:",nullzero"`

	AccessGroupMembers []*PGAccessGroupMemberEntry `bun:"rel:has-many,join:access_group_owner_public_key=access_group_owner_public_key,join:access_group_key_name=access_group_key_name"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `bun:",pk" pg:",pk,use_zero"`
}

type PGAccessGroupEntry struct {
	bun.BaseModel `bun:"table:access_group_entry"`
	AccessGroupEntry
}

type PGAccessGroupEntryUtxoOps struct {
	bun.BaseModel `bun:"table:access_group_entry_utxo_ops"`
	AccessGroupEntry
	UtxoOperation
}

// Convert the AccessGroup DeSo encoder to the PGAccessGroupEntry struct used by bun.
func AccessGroupEncoderToPGStruct(accessGroupEntry *lib.AccessGroupEntry, keyBytes []byte, params *lib.DeSoParams) AccessGroupEntry {
	pgAccessGroupEntry := AccessGroupEntry{
		ExtraData: consumer.ExtraDataBytesToString(accessGroupEntry.ExtraData),
		BadgerKey: keyBytes,
	}

	if accessGroupEntry.AccessGroupKeyName != nil {
		pgAccessGroupEntry.AccessGroupKeyName = string(accessGroupEntry.AccessGroupKeyName[:])
	}

	if accessGroupEntry.AccessGroupOwnerPublicKey != nil {
		pgAccessGroupEntry.AccessGroupOwnerPublicKey = consumer.PublicKeyBytesToBase58Check((*accessGroupEntry.AccessGroupOwnerPublicKey)[:], params)
	}

	if accessGroupEntry.AccessGroupPublicKey != nil {
		pgAccessGroupEntry.AccessGroupPublicKey = consumer.PublicKeyBytesToBase58Check((*accessGroupEntry.AccessGroupPublicKey)[:], params)
	}

	return pgAccessGroupEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func AccessGroupBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteAccessGroupEntry(entries, db, operationType)
	} else {
		err = bulkInsertAccessGroupEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertAccessGroupEntry inserts a batch of access_group entries into the database.
func bulkInsertAccessGroupEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGAccessGroupEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGAccessGroupEntry{AccessGroupEntry: AccessGroupEncoderToPGStruct(entry.Encoder.(*lib.AccessGroupEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertAccessGroupEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of access_group entries from the database.
func bulkDeleteAccessGroupEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGAccessGroupEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteAccessGroupEntry: Error deleting entries")
	}

	return nil
}
