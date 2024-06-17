package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// TODO: when to use nullzero vs use_zero?
type LockupYieldCurvePoint struct {
	ProfilePKID               string `bun:",nullzero"`
	LockupDurationNanoSecs    int64
	LockupYieldAPYBasisPoints uint64

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGLockupYieldCurvePoint struct {
	bun.BaseModel `bun:"table:yield_curve_point"`
	LockupYieldCurvePoint
}

// TODO: Do I need this?
type PGLockupYieldCurvePointUtxoOps struct {
	bun.BaseModel `bun:"table:yield_curve_point_utxo_ops"`
	LockupYieldCurvePoint
	UtxoOperation
}

// Convert the LockupYieldCurvePoint DeSo encoder to the PGLockedBalnceEntry struct used by bun.
func LockupYieldCurvePointEncoderToPGStruct(lockupYieldCurvePoint *lib.LockupYieldCurvePoint, keyBytes []byte, params *lib.DeSoParams) LockupYieldCurvePoint {
	pgLockupYieldCurvePoint := LockupYieldCurvePoint{
		BadgerKey: keyBytes,
	}

	if lockupYieldCurvePoint.ProfilePKID != nil {
		pgLockupYieldCurvePoint.ProfilePKID = consumer.PublicKeyBytesToBase58Check((*lockupYieldCurvePoint.ProfilePKID)[:], params)
	}

	pgLockupYieldCurvePoint.LockupDurationNanoSecs = lockupYieldCurvePoint.LockupDurationNanoSecs
	pgLockupYieldCurvePoint.LockupYieldAPYBasisPoints = lockupYieldCurvePoint.LockupYieldAPYBasisPoints

	return pgLockupYieldCurvePoint
}

// LockupYieldCurvePointBatchOperation is the entry point for processing a batch of LockedBalance entries.
// It determines the appropriate handler based on the operation type and executes it.
func LockupYieldCurvePointBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteLockupYieldCurvePoint(entries, db, operationType)
	} else {
		err = bulkInsertLockupYieldCurvePoint(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.LockupYieldCurvePointBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertLockupYieldCurvePoint inserts a batch of locked stake entries into the database.
func bulkInsertLockupYieldCurvePoint(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLockupYieldCurvePoint, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGLockupYieldCurvePoint{LockupYieldCurvePoint: LockupYieldCurvePointEncoderToPGStruct(entry.Encoder.(*lib.LockupYieldCurvePoint), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertLockupYieldCurvePoint: Error inserting entries")
	}
	return nil
}

// bulkDeleteLockupYieldCurvePoint deletes a batch of locked stake entries from the database.
func bulkDeleteLockupYieldCurvePoint(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLockupYieldCurvePoint{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteLockupYieldCurvePoint: Error deleting entries")
	}

	return nil
}
