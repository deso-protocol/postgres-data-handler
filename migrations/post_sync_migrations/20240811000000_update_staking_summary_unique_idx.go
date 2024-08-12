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
		drop materialized view validator_stats;
		drop materialized view staking_summary;
		
		create materialized view staking_summary as
		SELECT validator_summary.global_stake_amount_nanos,
			   validator_summary.num_validators,
			   current_epoch.current_epoch_number,
			   num_epochs_in_leader_schedule.num_epochs_in_leader_schedule,
			   staker_summary.num_stakers,
			   id
		FROM (SELECT sum(validator_entry.total_stake_amount_nanos)  AS global_stake_amount_nanos,
					 count(DISTINCT validator_entry.validator_pkid) AS num_validators
			  FROM validator_entry) validator_summary,
			 (SELECT max(epoch_entry.epoch_number) AS current_epoch_number
			  FROM epoch_entry) current_epoch,
			 (SELECT count(DISTINCT leader_schedule_entry.snapshot_at_epoch_number) AS num_epochs_in_leader_schedule
			  FROM leader_schedule_entry) num_epochs_in_leader_schedule,
			 (SELECT count(DISTINCT stake_entry.staker_pkid) AS num_stakers,
			1 AS id
			  FROM stake_entry) staker_summary;
		
		
		create unique index staking_summary_unique_index
			on staking_summary (id);
		
		create materialized view validator_stats as
		SELECT validator_entry.validator_pkid,
			   rank() OVER (ORDER BY (
				   CASE
					   WHEN validator_entry.jailed_at_epoch_number = 0 THEN 0
					   ELSE 1
					   END), validator_entry.total_stake_amount_nanos DESC, validator_entry.jailed_at_epoch_number DESC, validator_entry.validator_pkid) AS validator_rank,
			   validator_entry.total_stake_amount_nanos::double precision /
			   COALESCE(NULLIF(staking_summary.global_stake_amount_nanos::double precision, 0::double precision),
						1::double precision)                                                                                                             AS percent_total_stake,
			   COALESCE(jhe.time_in_jail, 0::numeric) +
			   CASE
				   WHEN validator_entry.jailed_at_epoch_number = 0 THEN 0::bigint
				   ELSE staking_summary.current_epoch_number - validator_entry.jailed_at_epoch_number
				   END::numeric                                                                                                                          AS epochs_in_jail,
			   COALESCE(leader_schedule_summary.num_epochs_in_leader_schedule,
						0::bigint)                                                                                                                       AS num_epochs_in_leader_schedule,
			   COALESCE(leader_schedule_summary.num_epochs_in_leader_schedule, 0::bigint)::double precision /
			   COALESCE(NULLIF(staking_summary.num_epochs_in_leader_schedule::double precision, 0::double precision),
						1::double precision)                                                                                                             AS percent_epochs_in_leader_schedule,
			   COALESCE(total_stake_rewards.total_rewards, 0::numeric)                                                                                   AS total_stake_reward_nanos
		FROM staking_summary,
			 validator_entry
				 LEFT JOIN (SELECT jhe_1.validator_pkid,
								   sum(jhe_1.unjailed_at_epoch_number - jhe_1.jailed_at_epoch_number) AS time_in_jail
							FROM jailed_history_event jhe_1
							GROUP BY jhe_1.validator_pkid) jhe ON jhe.validator_pkid::text = validator_entry.validator_pkid::text
				 LEFT JOIN (SELECT leader_schedule_entry.validator_pkid,
								   count(*) AS num_epochs_in_leader_schedule
							FROM leader_schedule_entry
							GROUP BY leader_schedule_entry.validator_pkid) leader_schedule_summary
						   ON leader_schedule_summary.validator_pkid::text = validator_entry.validator_pkid::text
				 LEFT JOIN (SELECT stake_reward.validator_pkid,
								   sum(stake_reward.reward_nanos) AS total_rewards
							FROM stake_reward
							GROUP BY stake_reward.validator_pkid) total_stake_rewards
						   ON total_stake_rewards.validator_pkid::text = validator_entry.validator_pkid::text;
		
		comment on materialized view validator_stats is E'@primaryKey validator_pkid\n@unique validator_rank\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStats|@fieldName validatorEntry';

		create unique index validator_stats_unique_index
			on validator_stats (validator_pkid);
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
			drop materialized view validator_stats;
			drop materialized view staking_summary;
			
			create materialized view staking_summary as
			SELECT validator_summary.global_stake_amount_nanos,
				   validator_summary.num_validators,
				   current_epoch.current_epoch_number,
				   num_epochs_in_leader_schedule.num_epochs_in_leader_schedule,
				   staker_summary.num_stakers
			FROM (SELECT sum(validator_entry.total_stake_amount_nanos)  AS global_stake_amount_nanos,
						 count(DISTINCT validator_entry.validator_pkid) AS num_validators
				  FROM validator_entry) validator_summary,
				 (SELECT max(epoch_entry.epoch_number) AS current_epoch_number
				  FROM epoch_entry) current_epoch,
				 (SELECT count(DISTINCT leader_schedule_entry.snapshot_at_epoch_number) AS num_epochs_in_leader_schedule
				  FROM leader_schedule_entry) num_epochs_in_leader_schedule,
				 (SELECT count(DISTINCT stake_entry.staker_pkid) AS num_stakers
				  FROM stake_entry) staker_summary;
			
			
			CREATE UNIQUE INDEX staking_summary_unique_index ON staking_summary (global_stake_amount_nanos, num_validators, current_epoch_number, num_epochs_in_leader_schedule);
			
			create materialized view validator_stats as
			SELECT validator_entry.validator_pkid,
				   rank() OVER (ORDER BY (
					   CASE
						   WHEN validator_entry.jailed_at_epoch_number = 0 THEN 0
						   ELSE 1
						   END), validator_entry.total_stake_amount_nanos DESC, validator_entry.jailed_at_epoch_number DESC, validator_entry.validator_pkid) AS validator_rank,
				   validator_entry.total_stake_amount_nanos::double precision /
				   COALESCE(NULLIF(staking_summary.global_stake_amount_nanos::double precision, 0::double precision),
							1::double precision)                                                                                                             AS percent_total_stake,
				   COALESCE(jhe.time_in_jail, 0::numeric) +
				   CASE
					   WHEN validator_entry.jailed_at_epoch_number = 0 THEN 0::bigint
					   ELSE staking_summary.current_epoch_number - validator_entry.jailed_at_epoch_number
					   END::numeric                                                                                                                          AS epochs_in_jail,
				   COALESCE(leader_schedule_summary.num_epochs_in_leader_schedule,
							0::bigint)                                                                                                                       AS num_epochs_in_leader_schedule,
				   COALESCE(leader_schedule_summary.num_epochs_in_leader_schedule, 0::bigint)::double precision /
				   COALESCE(NULLIF(staking_summary.num_epochs_in_leader_schedule::double precision, 0::double precision),
							1::double precision)                                                                                                             AS percent_epochs_in_leader_schedule,
				   COALESCE(total_stake_rewards.total_rewards, 0::numeric)                                                                                   AS total_stake_reward_nanos
			FROM staking_summary,
				 validator_entry
					 LEFT JOIN (SELECT jhe_1.validator_pkid,
									   sum(jhe_1.unjailed_at_epoch_number - jhe_1.jailed_at_epoch_number) AS time_in_jail
								FROM jailed_history_event jhe_1
								GROUP BY jhe_1.validator_pkid) jhe ON jhe.validator_pkid::text = validator_entry.validator_pkid::text
					 LEFT JOIN (SELECT leader_schedule_entry.validator_pkid,
									   count(*) AS num_epochs_in_leader_schedule
								FROM leader_schedule_entry
								GROUP BY leader_schedule_entry.validator_pkid) leader_schedule_summary
							   ON leader_schedule_summary.validator_pkid::text = validator_entry.validator_pkid::text
					 LEFT JOIN (SELECT stake_reward.validator_pkid,
									   sum(stake_reward.reward_nanos) AS total_rewards
								FROM stake_reward
								GROUP BY stake_reward.validator_pkid) total_stake_rewards
							   ON total_stake_rewards.validator_pkid::text = validator_entry.validator_pkid::text;
			
			comment on materialized view validator_stats is '@primaryKey validator_pkid
			@unique validator_rank
			@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStats|@fieldName validatorEntry';
			
			
			create unique index validator_stats_unique_index
				on validator_stats (validator_pkid);
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
