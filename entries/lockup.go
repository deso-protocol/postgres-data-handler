package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
)

// TODO: when to use nullzero vs use_zero?
type LockedBalanceEntry struct {
	HODLerPKID                  string `bun:",nullzero"`
	ProfilePKID                 string `bun:",nullzero"`
	UnlockTimestampNanoSecs     int64
	VestingEndTimestampNanoSecs int64
	BalanceBaseUnits            *bunbig.Int `pg:",use_zero"`

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGLockedBalanceEntry struct {
	bun.BaseModel `bun:"table:locked_balance_entry"`
	LockedBalanceEntry
}

// TODO: Do I need this?
type PGLockedBalanceEntryUtxoOps struct {
	bun.BaseModel `bun:"table:locked_balance_entry_utxo_ops"`
	LockedBalanceEntry
	UtxoOperation
}

// Convert the LockedBalanceEntry DeSo encoder to the PGLockedBalnceEntry struct used by bun.
func LockedBalanceEntryEncoderToPGStruct(lockedBalanceEntry *lib.LockedBalanceEntry, keyBytes []byte, params *lib.DeSoParams) LockedBalanceEntry {
	pgLockedBalanceEntry := LockedBalanceEntry{
		BadgerKey: keyBytes,
	}

	if lockedBalanceEntry.HODLerPKID != nil {
		pgLockedBalanceEntry.HODLerPKID = consumer.PublicKeyBytesToBase58Check((*lockedBalanceEntry.HODLerPKID)[:], params)
	}

	if lockedBalanceEntry.ProfilePKID != nil {
		pgLockedBalanceEntry.ProfilePKID = consumer.PublicKeyBytesToBase58Check((*lockedBalanceEntry.ProfilePKID)[:], params)
	}

	pgLockedBalanceEntry.UnlockTimestampNanoSecs = lockedBalanceEntry.UnlockTimestampNanoSecs
	pgLockedBalanceEntry.VestingEndTimestampNanoSecs = lockedBalanceEntry.VestingEndTimestampNanoSecs
	pgLockedBalanceEntry.BalanceBaseUnits = bunbig.FromMathBig(lockedBalanceEntry.BalanceBaseUnits.ToBig())

	return pgLockedBalanceEntry
}

// LockedBalanceEntryBatchOperation is the entry point for processing a batch of LockedBalance entries.
// It determines the appropriate handler based on the operation type and executes it.
func LockedBalanceEntryBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteLockedBalanceEntry(entries, db, operationType)
	} else {
		err = bulkInsertLockedBalanceEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.LockedBalanceEntryBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertLockedBalanceEntry inserts a batch of locked stake entries into the database.
func bulkInsertLockedBalanceEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLockedBalanceEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGLockedBalanceEntry{LockedBalanceEntry: LockedBalanceEntryEncoderToPGStruct(entry.Encoder.(*lib.LockedBalanceEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertLockedBalanceEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteLockedBalanceEntry deletes a batch of locked stake entries from the database.
func bulkDeleteLockedBalanceEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLockedBalanceEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteLockedBalanceEntry: Error deleting entries")
	}

	return nil
}
