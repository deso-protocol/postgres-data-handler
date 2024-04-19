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
		DROP MATERIALIZED VIEW IF EXISTS validator_stats;
create materialized view validator_stats as
select validator_entry.validator_pkid,
       rank()
       OVER ( order by (case when validator_entry.jailed_at_epoch_number = 0 then 0 else 1 end), validator_entry.total_stake_amount_nanos desc, validator_entry.jailed_at_epoch_number desc, validator_entry.validator_pkid) as validator_rank,
       validator_entry.total_stake_amount_nanos::float /
       coalesce(nullif(staking_summary.global_stake_amount_nanos::float, 0),
                1)                                                                                                                                                                                                           as percent_total_stake,
       coalesce(time_in_jail, 0) +
       (case
            when jailed_at_epoch_number = 0 then 0
            else (staking_summary.current_epoch_number - jailed_at_epoch_number) END)                                                                                                                                           epochs_in_jail,
       coalesce(leader_schedule_summary.num_epochs_in_leader_schedule, 0)                                                                                                                                                    as num_epochs_in_leader_schedule,
       coalesce(leader_schedule_summary.num_epochs_in_leader_schedule, 0)::float /
       coalesce(nullif(staking_summary.num_epochs_in_leader_schedule::float, 0),
                1)                                                                                                                                                                                                           as percent_epochs_in_leader_schedule,
       coalesce(total_rewards, 0)                                                                                                                                                                                            as total_stake_reward_nanos
from staking_summary,
     validator_entry
         left join (select validator_pkid, sum(jhe.unjailed_at_epoch_number - jhe.jailed_at_epoch_number) time_in_jail
                    from jailed_history_event jhe
                    group by validator_pkid) jhe
                   on jhe.validator_pkid = validator_entry.validator_pkid
         left join (select validator_pkid, count(*) as num_epochs_in_leader_schedule
                    from leader_schedule_entry
                    group by validator_pkid) leader_schedule_summary
                   on leader_schedule_summary.validator_pkid = validator_entry.validator_pkid
         left join (select validator_pkid, sum(reward_nanos) as total_rewards
                    from stake_reward
                    group by validator_pkid) as total_stake_rewards
                   on total_stake_rewards.validator_pkid = validator_entry.validator_pkid;
CREATE UNIQUE INDEX validator_stats_unique_index ON validator_stats (validator_pkid);
		comment on materialized view validator_stats is E'@primaryKey validator_pkid\n@unique validator_rank\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStats|@fieldName validatorEntry';
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
			DROP MATERIALIZED VIEW IF EXISTS validator_stats CASCADE;
CREATE MATERIALIZED VIEW validator_stats as
select validator_entry.validator_pkid,
       rank() OVER ( order by validator_entry.total_stake_amount_nanos) as            validator_rank,
       validator_entry.total_stake_amount_nanos::float /
       staking_summary.global_stake_amount_nanos::float                 as            percent_total_stake,
       coalesce(time_in_jail, 0) +
       (case
            when jailed_at_epoch_number = 0 then 0
            else (staking_summary.current_epoch_number - jailed_at_epoch_number) END) epochs_in_jail,
       coalesce(leader_schedule_summary.num_epochs_in_leader_schedule, 0) num_epochs_in_leader_schedule,
       coalesce(leader_schedule_summary.num_epochs_in_leader_schedule, 0)::float /
       staking_summary.num_epochs_in_leader_schedule::float             as            percent_epochs_in_leader_schedule,
       coalesce(total_rewards, 0)                                       as            total_stake_reward_nanos
from staking_summary,
     validator_entry
         left join (select validator_pkid, sum(jhe.unjailed_at_epoch_number - jhe.jailed_at_epoch_number) time_in_jail
                    from jailed_history_event jhe
                    group by validator_pkid) jhe
                   on jhe.validator_pkid = validator_entry.validator_pkid
         left join (select validator_pkid, count(*) as num_epochs_in_leader_schedule
                    from leader_schedule_entry
                    group by validator_pkid) leader_schedule_summary
                   on leader_schedule_summary.validator_pkid = validator_entry.validator_pkid
         left join (select validator_pkid, sum(reward_nanos) as total_rewards
                    from stake_reward
                    group by validator_pkid) as total_stake_rewards
                   on total_stake_rewards.validator_pkid = validator_entry.validator_pkid;

CREATE UNIQUE INDEX validator_stats_unique_index ON validator_stats (validator_pkid);

		comment on materialized view validator_stats is E'@unique validator_pkid\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStats|@fieldName validatorEntry';
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
