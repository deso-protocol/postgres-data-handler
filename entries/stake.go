package entries

import (
	"context"
	"github.com/deso-protocol/backend/routes"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
)

// TODO: when to use nullzero vs use_zero?
type StakeEntry struct {
	StakerPKID       string `bun:",nullzero"`
	ValidatorPKID    string `bun:",nullzero"`
	RewardMethod     routes.StakeRewardMethod
	StakeAmountNanos *bunbig.Int `pg:",use_zero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGStakeEntry struct {
	bun.BaseModel `bun:"table:stake_entry"`
	StakeEntry
}

// TODO: Do I need this?
type PGStakeEntryUtxoOps struct {
	bun.BaseModel `bun:"table:stake_entry_utxo_ops"`
	StakeEntry
	UtxoOperation
}

// Convert the StakeEntry DeSo encoder to the PGStakeEntry struct used by bun.
func StakeEncoderToPGStruct(stakeEntry *lib.StakeEntry, keyBytes []byte, params *lib.DeSoParams) StakeEntry {
	pgStakeEntry := StakeEntry{
		ExtraData: consumer.ExtraDataBytesToString(stakeEntry.ExtraData),
		BadgerKey: keyBytes,
	}

	if stakeEntry.StakerPKID != nil {
		pgStakeEntry.StakerPKID = consumer.PublicKeyBytesToBase58Check((*stakeEntry.StakerPKID)[:], params)
	}

	if stakeEntry.ValidatorPKID != nil {
		pgStakeEntry.ValidatorPKID = consumer.PublicKeyBytesToBase58Check((*stakeEntry.ValidatorPKID)[:], params)
	}

	pgStakeEntry.RewardMethod = routes.FromLibStakeRewardMethod(stakeEntry.RewardMethod)
	pgStakeEntry.StakeAmountNanos = bunbig.FromMathBig(stakeEntry.StakeAmountNanos.ToBig())

	return pgStakeEntry
}

// StakeBatchOperation is the entry point for processing a batch of Stake entries.
// It determines the appropriate handler based on the operation type and executes it.
func StakeBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteStakeEntry(entries, db, operationType)
	} else {
		err = bulkInsertStakeEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.StakeBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertStakeEntry inserts a batch of stake entries into the database.
func bulkInsertStakeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGStakeEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGStakeEntry{StakeEntry: StakeEncoderToPGStruct(entry.Encoder.(*lib.StakeEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertStakeEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteStakeEntry deletes a batch of stake entries from the database.
func bulkDeleteStakeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGStakeEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteStakeEntry: Error deleting entries")
	}

	return nil
}
