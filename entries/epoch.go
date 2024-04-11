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
	InitialLeaderIndexOffset        uint64
	CreatedAtBlockTimestampNanoSecs int64
	SnapshotAtEpochNumber           uint64
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

	var snapshotAtEpochNumber uint64
	// Epochs use data snapshotted from two epochs ago. Epochs 0 and 1 use data from epoch 0.
	if epochEntry.EpochNumber >= 2 {
		snapshotAtEpochNumber = epochEntry.EpochNumber - 2
	}
	return EpochEntry{
		EpochNumber:                     epochEntry.EpochNumber,
		InitialBlockHeight:              epochEntry.InitialBlockHeight,
		InitialView:                     epochEntry.InitialView,
		FinalBlockHeight:                epochEntry.FinalBlockHeight,
		InitialLeaderIndexOffset:        epochEntry.InitialLeaderIndexOffset,
		CreatedAtBlockTimestampNanoSecs: epochEntry.CreatedAtBlockTimestampNanoSecs,
		SnapshotAtEpochNumber:           snapshotAtEpochNumber,
	}
}

// EpochEntryBatchOperation is the entry point for processing a batch of Epoch entries.
// It determines the appropriate handler based on the operation type and executes it.
func EpochEntryBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	// Core only tracks the current epoch entry and never deletes them.
	// In order to track all historical epoch entries, we don't use the badger
	// key to uniquely identify them, but rather the epoch number.
	if operationType == lib.DbOperationTypeDelete {
		return errors.Wrapf(err, "entries.EpochEntryBatchOperation: Delete operation type not supported")
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
		query = query.On("CONFLICT (epoch_number) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertEpochEntry: Error inserting entries")
	}
	return nil
}
