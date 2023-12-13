package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// TODO: when to use nullzero vs use_zero?
type EpochEntry struct {
	EpochNumber                     uint64
	InitialBlockHeight              uint64
	InitialView                     uint64
	FinalBlockHeight                uint64
	CreatedAtBlockTimestampNanoSecs uint64

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGEpochEntry struct {
	bun.BaseModel `bun:"table:epoch_entry"`
	EpochEntry
}

// TODO: Do I need this?
type PGEpochUtxoOps struct {
	bun.BaseModel `bun:"table:epoch_entry_utxo_ops"`
	EpochEntry
	UtxoOperation
}

// Convert the EpochEntry DeSo encoder to the PGEpochEntry struct used by bun.
func EpochEntryEncoderToPGStruct(epochEntry *lib.EpochEntry, keyBytes []byte, params *lib.DeSoParams) EpochEntry {
	return EpochEntry{
		EpochNumber:        epochEntry.EpochNumber,
		InitialBlockHeight: epochEntry.InitialBlockHeight,
		InitialView:        epochEntry.InitialView,
		FinalBlockHeight:   epochEntry.FinalBlockHeight,
		BadgerKey:          keyBytes,
	}
}

// EpochEntryBatchOperation is the entry point for processing a batch of Epoch entries.
// It determines the appropriate handler based on the operation type and executes it.
func EpochEntryBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteEpochEntry(entries, db, operationType)
	} else {
		err = bulkInsertEpochEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.EpochEntryBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertEpochEntry inserts a batch of locked stake entries into the database.
func bulkInsertEpochEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGEpochEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGEpochEntry{EpochEntry: EpochEntryEncoderToPGStruct(entry.Encoder.(*lib.EpochEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertEpochEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteEpochEntry deletes a batch of locked stake entries from the database.
func bulkDeleteEpochEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGEpochEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteEpochEntry: Error deleting entries")
	}

	return nil
}
