package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		if !calculateExplorerStatistics {
			return nil
		}
		_, err := db.Exec(`
		CREATE OR REPLACE VIEW stake_rewards_by_validator_and_staker AS
		SELECT
			stake_reward.validator_pkid,
			stake_reward.staker_pkid,
			SUM(stake_reward.reward_nanos) AS reward_nanos
		FROM stake_reward
		GROUP BY stake_reward.validator_pkid, stake_reward.staker_pkid;
		comment on view stake_rewards_by_validator_and_staker is E'@primaryKey staker_pkid,validator_pkid\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStakeRewardsByStaker|@fieldName validatorEntry\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName accountStakeRewardsByValidator|@fieldName validatorAccount\n@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName accountStakeRewardsByStaker|@fieldName stakerAccount';
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		if !calculateExplorerStatistics {
			return nil
		}
		_, err := db.Exec(`
			DROP VIEW IF EXISTS stake_rewards_by_validator_and_staker CASCADE;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
