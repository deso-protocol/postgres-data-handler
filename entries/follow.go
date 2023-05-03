package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGFollowEntry struct {
	bun.BaseModel `bun:"table:follow_entry"`
	FollowerPkid  string `pg:",use_zero"`
	FollowedPkid  string `pg:",use_zero"`
	BadgerKey     []byte `pg:",pk,use_zero"`
}

// Convert the follow DeSo encoder to the PG struct used by bun.
func FollowEncoderToPGStruct(followEntry *lib.FollowEntry, keyBytes []byte) *PGFollowEntry {
	return &PGFollowEntry{
		FollowerPkid: consumer.PublicKeyBytesToBase58Check(followEntry.FollowerPKID[:]),
		FollowedPkid: consumer.PublicKeyBytesToBase58Check(followEntry.FollowedPKID[:]),
		BadgerKey:    keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func FollowBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteFollowEntry(entries, db, operationType)
	} else {
		err = bulkInsertFollowEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertFollowEntry inserts a batch of follow entries into the database.
func bulkInsertFollowEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGFollowEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = FollowEncoderToPGStruct(entry.Encoder.(*lib.FollowEntry), entry.KeyBytes)
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertFollowEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of follow entries from the database.
func bulkDeleteFollowEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGFollowEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteFollowEntry: Error deleting entries")
	}

	return nil
}