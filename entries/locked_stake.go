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
type LockedStakeEntry struct {
	StakerPKID          string      `bun:",nullzero"`
	ValidatorPKID       string      `bun:",nullzero"`
	LockedAmountNanos   *bunbig.Int `pg:",use_zero"`
	LockedAtEpochNumber uint64

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGLockedStakeEntry struct {
	bun.BaseModel `bun:"table:locked_stake_entry"`
	LockedStakeEntry
}

// TODO: Do I need this?
type PGLockedStakeEntryUtxoOps struct {
	bun.BaseModel `bun:"table:locked_stake_entry_utxo_ops"`
	LockedStakeEntry
	UtxoOperation
}

// Convert the LockedStakeEntry DeSo encoder to the PGLockedStakeEntry struct used by bun.
func LockedStakeEncoderToPGStruct(lockedStakeEntry *lib.LockedStakeEntry, keyBytes []byte, params *lib.DeSoParams) LockedStakeEntry {
	pgLockedStakeEntry := LockedStakeEntry{
		ExtraData: consumer.ExtraDataBytesToString(lockedStakeEntry.ExtraData, params),
		BadgerKey: keyBytes,
	}

	if lockedStakeEntry.StakerPKID != nil {
		pgLockedStakeEntry.StakerPKID = consumer.PublicKeyBytesToBase58Check((*lockedStakeEntry.StakerPKID)[:], params)
	}

	if lockedStakeEntry.ValidatorPKID != nil {
		pgLockedStakeEntry.ValidatorPKID = consumer.PublicKeyBytesToBase58Check((*lockedStakeEntry.ValidatorPKID)[:], params)
	}

	pgLockedStakeEntry.LockedAtEpochNumber = lockedStakeEntry.LockedAtEpochNumber
	pgLockedStakeEntry.LockedAmountNanos = bunbig.FromMathBig(lockedStakeEntry.LockedAmountNanos.ToBig())

	return pgLockedStakeEntry
}

// LockedStakeBatchOperation is the entry point for processing a batch of LockedStake entries.
// It determines the appropriate handler based on the operation type and executes it.
func LockedStakeBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteLockedStakeEntry(entries, db, operationType)
	} else {
		err = bulkInsertLockedStakeEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.LockedStakeBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertLockedStakeEntry inserts a batch of locked stake entries into the database.
func bulkInsertLockedStakeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLockedStakeEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGLockedStakeEntry{LockedStakeEntry: LockedStakeEncoderToPGStruct(entry.Encoder.(*lib.LockedStakeEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertLockedStakeEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteLockedStakeEntry deletes a batch of locked stake entries from the database.
func bulkDeleteLockedStakeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLockedStakeEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteLockedStakeEntry: Error deleting entries")
	}

	return nil
}
