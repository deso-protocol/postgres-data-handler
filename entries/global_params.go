package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type GlobalParamsEntry struct {
	USDCentsPerBitcoin                             uint64
	CreateProfileFeeNanos                          uint64
	CreateNFTFeeNanos                              uint64
	MaxCopiesPerNFT                                uint64
	MinimumNetworkFeeNanosPerKB                    uint64
	MaxNonceExpirationBlockHeightOffset            uint64
	StakeLockupEpochDuration                       uint64
	ValidatorJailEpochDuration                     uint64
	LeaderScheduleMaxNumValidators                 uint64
	ValidatorSetMaxNumValidators                   uint64
	StakingRewardsMaxNumStakes                     uint64
	StakingRewardsAPYBasisPoints                   uint64
	EpochDurationNumBlocks                         uint64
	JailInactiveValidatorGracePeriodEpochs         uint64
	MaximumVestedIntersectionsPerLockupTransaction int
	FeeBucketGrowthRateBasisPoints                 uint64
	BlockTimestampDriftNanoSecs                    int64
	MempoolMaxSizeBytes                            uint64
	MempoolFeeEstimatorNumMempoolBlocks            uint64
	MempoolFeeEstimatorNumPastBlocks               uint64
	MaxBlockSizeBytesPoS                           uint64
	SoftMaxBlockSizeBytesPoS                       uint64
	MaxTxnSizeBytesPoS                             uint64
	BlockProductionIntervalMillisecondsPoS         uint64
	TimeoutIntervalMillisecondsPoS                 uint64

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGGlobalParamsEntry struct {
	bun.BaseModel `bun:"table:global_params_entry"`
	GlobalParamsEntry
}

// Convert the GlobalParams DeSo encoder to the PGGlobalParamsEntry struct used by bun.
func GlobalParamsEncoderToPGStruct(globalParamsEntry *lib.GlobalParamsEntry, keyBytes []byte, params *lib.DeSoParams) GlobalParamsEntry {
	mergedGlobalParamsEntry := lib.MergeGlobalParamEntryDefaults(globalParamsEntry, params)

	return GlobalParamsEntry{
		USDCentsPerBitcoin:                             mergedGlobalParamsEntry.USDCentsPerBitcoin,
		CreateProfileFeeNanos:                          mergedGlobalParamsEntry.CreateProfileFeeNanos,
		CreateNFTFeeNanos:                              mergedGlobalParamsEntry.CreateNFTFeeNanos,
		MaxCopiesPerNFT:                                mergedGlobalParamsEntry.MaxCopiesPerNFT,
		MinimumNetworkFeeNanosPerKB:                    mergedGlobalParamsEntry.MinimumNetworkFeeNanosPerKB,
		MaxNonceExpirationBlockHeightOffset:            mergedGlobalParamsEntry.MaxNonceExpirationBlockHeightOffset,
		StakeLockupEpochDuration:                       mergedGlobalParamsEntry.StakeLockupEpochDuration,
		ValidatorJailEpochDuration:                     mergedGlobalParamsEntry.ValidatorJailEpochDuration,
		LeaderScheduleMaxNumValidators:                 mergedGlobalParamsEntry.LeaderScheduleMaxNumValidators,
		ValidatorSetMaxNumValidators:                   mergedGlobalParamsEntry.ValidatorSetMaxNumValidators,
		StakingRewardsMaxNumStakes:                     mergedGlobalParamsEntry.StakingRewardsMaxNumStakes,
		StakingRewardsAPYBasisPoints:                   mergedGlobalParamsEntry.StakingRewardsAPYBasisPoints,
		EpochDurationNumBlocks:                         mergedGlobalParamsEntry.EpochDurationNumBlocks,
		JailInactiveValidatorGracePeriodEpochs:         mergedGlobalParamsEntry.JailInactiveValidatorGracePeriodEpochs,
		MaximumVestedIntersectionsPerLockupTransaction: mergedGlobalParamsEntry.MaximumVestedIntersectionsPerLockupTransaction,
		FeeBucketGrowthRateBasisPoints:                 mergedGlobalParamsEntry.FeeBucketGrowthRateBasisPoints,
		BlockTimestampDriftNanoSecs:                    mergedGlobalParamsEntry.BlockTimestampDriftNanoSecs,
		MempoolMaxSizeBytes:                            mergedGlobalParamsEntry.MempoolMaxSizeBytes,
		MempoolFeeEstimatorNumMempoolBlocks:            mergedGlobalParamsEntry.MempoolFeeEstimatorNumMempoolBlocks,
		MempoolFeeEstimatorNumPastBlocks:               mergedGlobalParamsEntry.MempoolFeeEstimatorNumPastBlocks,
		MaxBlockSizeBytesPoS:                           mergedGlobalParamsEntry.MaxBlockSizeBytesPoS,
		SoftMaxBlockSizeBytesPoS:                       mergedGlobalParamsEntry.SoftMaxBlockSizeBytesPoS,
		MaxTxnSizeBytesPoS:                             mergedGlobalParamsEntry.MaxTxnSizeBytesPoS,
		BlockProductionIntervalMillisecondsPoS:         mergedGlobalParamsEntry.BlockProductionIntervalMillisecondsPoS,
		TimeoutIntervalMillisecondsPoS:                 mergedGlobalParamsEntry.TimeoutIntervalMillisecondsPoS,

		BadgerKey: keyBytes,
	}
}

// GlobalParamsBatchOperation is the entry point for processing a batch of global params entries.
// It determines the appropriate handler based on the operation type and executes it.
func GlobalParamsBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteGlobalParamsEntry(entries, db, operationType)
	} else {
		err = bulkInsertGlobalParamsEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertGlobalParamsEntry inserts a batch of global_params entries into the database.
func bulkInsertGlobalParamsEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGGlobalParamsEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGGlobalParamsEntry{GlobalParamsEntry: GlobalParamsEncoderToPGStruct(entry.Encoder.(*lib.GlobalParamsEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertGlobalParamsEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of global_params entries from the database.
func bulkDeleteGlobalParamsEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGGlobalParamsEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteGlobalParamsEntry: Error deleting entries")
	}

	return nil
}
