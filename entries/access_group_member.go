package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGAccessGroupMemberEntry struct {
	bun.BaseModel              `bun:"table:access_group_member_entry"`
	AccessGroupMemberPublicKey string `pg:",use_zero"`
	AccessGroupMemberKeyName   string `pg:",use_zero"`
	EncryptedKey               []byte `pg:",use_zero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

// Convert the AccessGroupMember DeSo encoder to the PGAccessGroupMemberEntry struct used by bun.
func AccessGroupMemberEncoderToPGStruct(accessGroupMemberEntry *lib.AccessGroupMemberEntry, keyBytes []byte) *PGAccessGroupMemberEntry {
	pgAccessGroupMemberEntry := &PGAccessGroupMemberEntry{
		EncryptedKey: accessGroupMemberEntry.EncryptedKey,
		ExtraData:    consumer.ExtraDataBytesToString(accessGroupMemberEntry.ExtraData),
		BadgerKey:    keyBytes,
	}

	if accessGroupMemberEntry.AccessGroupMemberKeyName != nil {
		pgAccessGroupMemberEntry.AccessGroupMemberKeyName = string(accessGroupMemberEntry.AccessGroupMemberKeyName[:])
	}

	if accessGroupMemberEntry.AccessGroupMemberPublicKey != nil {
		pgAccessGroupMemberEntry.AccessGroupMemberPublicKey = consumer.PublicKeyBytesToBase58Check((*accessGroupMemberEntry.AccessGroupMemberPublicKey)[:])
	}

	return pgAccessGroupMemberEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func AccessGroupMemberBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteAccessGroupMemberEntry(entries, db, operationType)
	} else {
		err = bulkInsertAccessGroupMemberEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertAccessGroupMemberEntry inserts a batch of access_group_member entries into the database.
func bulkInsertAccessGroupMemberEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGAccessGroupMemberEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = AccessGroupMemberEncoderToPGStruct(entry.Encoder.(*lib.AccessGroupMemberEntry), entry.KeyBytes)
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertAccessGroupMemberEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of access_group_member entries from the database.
func bulkDeleteAccessGroupMemberEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGAccessGroupMemberEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteAccessGroupMemberEntry: Error deleting entries")
	}

	return nil
}