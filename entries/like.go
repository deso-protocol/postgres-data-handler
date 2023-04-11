package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGLikeEntry struct {
	bun.BaseModel `bun:"table:like_entry"`
	PublicKey     string `pg:",use_zero" decode_function:"base_58_check" decode_src_field_name:"LikerPubKey"`
	PostHash      string `pg:",use_zero" decode_function:"blockhash" decode_src_field_name:"LikedPostHash"`
	// This is the PKID
	BadgerKey []byte `pg:",pk,use_zero"`
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func LikeBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteLikeEntry(entries, db, operationType)
	} else {
		err = bulkInsertLikeEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertLikeEntry inserts a batch of like entries into the database.
func bulkInsertLikeEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLikeEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for i := len(uniqueEntries) - 1; i >= 0; i-- {
		encoder := uniqueEntries[i].Encoder
		pgProfileEntry := &PGLikeEntry{}
		// Copy all encoder fields to the bun struct.
		consumer.CopyStruct(encoder, pgProfileEntry)
		// Add the badger key to the struct.
		pgProfileEntry.BadgerKey = entries[i].KeyBytes
		pgEntrySlice[i] = pgProfileEntry
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertLikeEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of like entries from the database.
func bulkDeleteLikeEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLikeEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteLikeEntry: Error deleting entries")
	}

	return nil
}
