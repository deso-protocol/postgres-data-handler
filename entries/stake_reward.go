package entries

import (
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/uptrace/bun"
)

type StakeReward struct {
	StakerPKID            string                  `bun:",nullzero"`
	ValidatorPKID         string                  `bun:",nullzero"`
	RewardMethod          lib.StakingRewardMethod // TODO: we probably want this to be human readable?
	RewardNanos           uint64                  `pg:",use_zero"`
	IsValidatorCommission bool
	BlockHash             string

	UtxoOpIndex uint64 `pg:",use_zero"`
}

type PGStakeReward struct {
	bun.BaseModel `bun:"table:stake_reward"`
	StakeReward
}

// Convert the StakeRewardStateChangeMetadata DeSo encoder to the PGStakeReward struct used by bun.
func StakeRewardEncoderToPGStruct(
	stakeReward *lib.StakeRewardStateChangeMetadata,
	params *lib.DeSoParams,
	blockHash string,
	utxoOpIndex uint64,
) StakeReward {
	pgStakeReward := StakeReward{}

	if stakeReward.StakerPKID != nil {
		pgStakeReward.StakerPKID = consumer.PublicKeyBytesToBase58Check((*stakeReward.StakerPKID)[:], params)
	}

	if stakeReward.ValidatorPKID != nil {
		pgStakeReward.ValidatorPKID = consumer.PublicKeyBytesToBase58Check((*stakeReward.ValidatorPKID)[:], params)
	}

	pgStakeReward.RewardMethod = stakeReward.StakingRewardMethod
	pgStakeReward.RewardNanos = stakeReward.RewardNanos
	pgStakeReward.IsValidatorCommission = stakeReward.IsValidatorCommission
	pgStakeReward.BlockHash = blockHash
	pgStakeReward.UtxoOpIndex = utxoOpIndex
	return pgStakeReward
}
