package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

// TODO: revisit access group relationships when we refactor the messaging app to use the graphql API.
func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
                comment on table stake_reward is E'@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName stakeRewards|@fieldName staker\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName validatorStakeRewards|@fieldName validator\n@foreignKey (block_hash) references block (block_hash)|@foreignFieldName stakeRewardForBlock|@fieldName block';
                comment on table stake_entry is E'@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName stakeEntries|@fieldName staker\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName validatorStakeEntries|@fieldName validatorAccount\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName stakeEntries|@fieldName validatorEntry';
                comment on table validator_entry is E'@unique validator_pkid\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName validatorEntry|@fieldName account';
                comment on table locked_stake_entry is E'@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName lockedStakeEntries|@fieldName staker\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName validatorLockedStakeEntries|@fieldName validatorAccount\n@foreignKey (validator_pkid) references validator_entry(validator_pkid)|@foreignFieldName validatorLockedStakeEntries|@fieldName validatorEntry';
                comment on table yield_curve_point is E'@foreignKey (profile_pkid) references account (pkid)|@foreignFieldName yieldCurvePoints|@fieldName account';
                comment on table locked_balance_entry is E'@foreignKey (profile_pkid) references account (pkid)|@foreignFieldName profileLockedBalanceEntries|@fieldName profileAccount\n@foreignKey (hodler_pkid) references account (pkid)|@foreignFieldName hodlerLockedBalanceEntries|@fieldName hodlerAccount';
                comment on table block is E'@unique block_hash\n@unique height\n@foreignKey (proposer_public_key) references account (public_key)|@foreignFieldName blocksProposed|@fieldName blockProposer';
                comment on column stake_entry.badger_key is E'@omit';
                comment on column validator_entry.badger_key is E'@omit';
                comment on column locked_stake_entry.badger_key is E'@omit';
                comment on column yield_curve_point.badger_key is E'@omit';
                comment on column locked_balance_entry.badger_key is E'@omit';
                comment on table transaction_partition_34 is E'@omit';
                comment on table transaction_partition_35 is E'@omit';
                comment on table transaction_partition_36 is E'@omit';
                comment on table transaction_partition_37 is E'@omit';
                comment on table transaction_partition_38 is E'@omit';
                comment on table transaction_partition_39 is E'@omit';
                comment on table transaction_partition_40 is E'@omit';
                comment on table transaction_partition_41 is E'@omit';
                comment on table transaction_partition_42 is E'@omit';
                comment on table transaction_partition_43 is E'@omit';
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
				comment on table stake_reward is NULL;
                comment on table stake_entry is NULL;
                comment on table validator_entry is NULL;
                comment on table locked_stake_entry is NULL;
                comment on table yield_curve_point is NULL;
                comment on table locked_balance_entry is NULL;
                comment on table block is E'@unique block_hash\n@unique height';
                comment on column stake_entry.badger_key is NULL;
                comment on column validator_entry.badger_key is NULL;
                comment on column locked_stake_entry.badger_key is NULL;
                comment on column yield_curve_point.badger_key is NULL;
                comment on column locked_balance_entry.badger_key is NULL;
                comment on column epoch_entry.badger_key is NULL;
                comment on table transaction_partition_34 is NULL;
                comment on table transaction_partition_35 is NULL;
                comment on table transaction_partition_36 is NULL;
                comment on table transaction_partition_37 is NULL;
                comment on table transaction_partition_38 is NULL;
                comment on table transaction_partition_39 is NULL;
                comment on table transaction_partition_40 is NULL;
                comment on table transaction_partition_41 is NULL;
                comment on table transaction_partition_42 is NULL;
                comment on table transaction_partition_43 is NULL;	
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
