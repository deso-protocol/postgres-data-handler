package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGDiamondEntry struct {
	bun.BaseModel `bun:"table:diamond_entry"`
	SenderPkid    []byte `pg:",use_zero" decode_function:"pkid" decode_src_field_name:"SenderPKID"`
	ReceiverPkid  []byte `pg:",use_zero" decode_function:"pkid" decode_src_field_name:"ReceiverPKID"`
	DiamondLevel  int64  `pg:",use_zero"`
	PostHash      string `pg:",use_zero" decode_function:"blockhash" decode_src_field_name:"DiamondPostHash"`
	BadgerKey     []byte `pg:",pk,use_zero"`
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func DiamondBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteDiamondEntry(entries, db, operationType)
	} else {
		err = bulkInsertDiamondEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertDiamondEntry inserts a batch of diamond entries into the database.
func bulkInsertDiamondEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGDiamondEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for i := len(uniqueEntries) - 1; i >= 0; i-- {
		encoder := uniqueEntries[i].Encoder
		pgEntry := &PGDiamondEntry{}
		// Copy all encoder fields to the bun struct.
		consumer.CopyStruct(encoder, pgEntry)
		// Add the badger key to the struct.
		pgEntry.BadgerKey = entries[i].KeyBytes
		pgEntrySlice[i] = pgEntry
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertDiamondEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of diamond entries from the database.
func bulkDeleteDiamondEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGDiamondEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteDiamondEntry: Error deleting entries")
	}

	return nil
}
