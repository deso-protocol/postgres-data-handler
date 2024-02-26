package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`

CREATE OR REPLACE VIEW epoch_details_for_block as
select block_hash, epoch_number, bls.pkid as proposer_pkid
from block
         left join epoch_entry
                   on epoch_entry.initial_block_height <= block.height and
                      epoch_entry.final_block_height >= block.height
         left join bls_public_key_pkid_pair_snapshot_entry bls
                   on bls.snapshot_at_epoch_number = epoch_entry.snapshot_at_epoch_number and
                      block.proposer_voting_public_key = bls.bls_public_key;

		comment on view epoch_details_for_block is E'@unique block_hash\n@unique epoch_number\n@foreignKey (block_hash) references block (block_hash)|@foreignFieldName epochDetailForBlock|@fieldName block\n@foreignKey (epoch_number) references epoch_entry (epoch_number)|@foreignFieldName blockHashesInEpoch|@fieldName epochEntry\n@foreignKey (proposer_pkid) references account (pkid)|@foreignFieldName proposedBlockHashes|@fieldName proposer';
		comment on table bls_public_key_pkid_pair_snapshot_entry is E'@foreignKey (pkid) references account (pkid)|@foreignFieldName blsPublicKeyPkidPairSnapshotEntries|@fieldName account\n@foreignKey (snapshot_at_epoch_number) references epoch_entry (snapshot_at_epoch_number)|@foreignFieldName blsPublicKeyPkidPairSnapshotEntries|@fieldName epochEntry';
		comment on column bls_public_key_pkid_pair_snapshot_entry.badger_key is E'@omit';
`)
		if err != nil {
			return err
		}
		if !calculateExplorerStatistics {
			return nil
		}
		_, err = db.Exec(`
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

CREATE MATERIALIZED VIEW staking_summary as
select *
from (select sum(total_stake_amount_nanos)  as global_stake_amount_nanos,
             count(distinct validator_pkid) as num_validators
      from validator_entry) validator_summary,
     (select max(epoch_number) current_epoch_number from epoch_entry) current_epoch,
     (select count(distinct snapshot_at_epoch_number) num_epochs_in_leader_schedule
      from leader_schedule_entry) num_epochs_in_leader_schedule;

CREATE UNIQUE INDEX staking_summary_unique_index ON staking_summary (global_stake_amount_nanos, num_validators, current_epoch_number, num_epochs_in_leader_schedule);

CREATE MATERIALIZED VIEW validator_stats as
select validator_entry.validator_pkid,
       rank() OVER ( order by validator_entry.total_stake_amount_nanos) as            validator_rank,
       validator_entry.total_stake_amount_nanos::float /
       staking_summary.global_stake_amount_nanos::float                 as            percent_total_stake,
       coalesce(time_in_jail, 0) +
       (case
            when jailed_at_epoch_number = 0 then 0
            else (staking_summary.current_epoch_number - jailed_at_epoch_number) END) epochs_in_jail,
       coalesce(leader_schedule_summary.num_epochs_in_leader_schedule, 0),
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
		comment on materialized view my_stake_summary is E'@foreignKey (staker_pkid) references account (pkid)|@foreignFieldName myStakeSummary|@fieldName staker';

`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`	
			comment on column bls_public_key_pkid_pair_snapshot_entry.badger_key is NULL;	
			comment on table bls_public_key_pkid_pair_snapshot_entry is NULL; 
			DROP VIEW IF EXISTS epoch_details_for_block CASCADE;
`)
		if err != nil {
			return err
		}
		if !calculateExplorerStatistics {
			return nil
		}
		_, err = db.Exec(`
			DROP MATERIALIZED VIEW IF EXISTS validator_stats CASCADE;
			DROP MATERIALIZED VIEW IF EXISTS staking_summary CASCADE;
			DROP MATERIALIZED VIEW IF EXISTS my_stake_summary CASCADE;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
