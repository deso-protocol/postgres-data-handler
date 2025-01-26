package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
DROP MATERIALIZED VIEW if exists my_stake_summary;

CREATE MATERIALIZED VIEW my_stake_summary as
with staker_pkids as (select staker_pkid
                      from stake_reward
                      union
                      select staker_pkid
                      from stake_entry
                      union
                      select staker_pkid
                      from locked_stake_entry)
select staker_pkids.staker_pkid        as staker_pkid,
       coalesce(total_rewards, 0)      as total_stake_rewards,
       coalesce(total_stake, 0)        as total_stake,
       coalesce(total_locked_stake, 0) as total_locked_stake
from staker_pkids
         left join (select staker_pkid, sum(reward_nanos) total_rewards
                    from stake_reward
                    group by staker_pkid) total_stake_rewards
                   on total_stake_rewards.staker_pkid = staker_pkids.staker_pkid
         left join (select staker_pkid, sum(stake_amount_nanos) total_stake
                    from stake_entry
                    group by staker_pkid) total_stake_amount
                   on total_stake_amount.staker_pkid = staker_pkids.staker_pkid
         left join (select staker_pkid, sum(locked_stake_entry.locked_amount_nanos) total_locked_stake
                    from locked_stake_entry
                    group by staker_pkid) total_locked_stake
                   on total_locked_stake.staker_pkid = staker_pkids.staker_pkid;

CREATE UNIQUE INDEX my_stake_summary_unique_index ON my_stake_summary (staker_pkid);

comment on materialized view my_stake_summary is E'@unique staker_pkid\n@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName myStakeSummary|@fieldName staker';
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
DROP MATERIALIZED VIEW if exists my_stake_summary;

CREATE MATERIALIZED VIEW my_stake_summary as
select coalesce(total_stake_rewards.staker_pkid, total_stake_amount.staker_pkid) as staker_pkid,
       total_stake_rewards.total_rewards                                         as total_stake_rewards,
       total_stake_amount.total_stake                                            as total_stake
from (select staker_pkid, sum(reward_nanos) total_rewards
      from stake_reward
      group by staker_pkid) total_stake_rewards
         full outer join
     (select staker_pkid, sum(stake_amount_nanos) total_stake
      from stake_entry
      group by staker_pkid) total_stake_amount
     on total_stake_amount.staker_pkid = total_stake_rewards.staker_pkid;

CREATE UNIQUE INDEX my_stake_summary_unique_index ON my_stake_summary (staker_pkid);

comment on materialized view my_stake_summary is E'@unique staker_pkid\n@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName myStakeSummary|@fieldName staker';
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
