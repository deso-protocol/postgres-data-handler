package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
CREATE OR REPLACE VIEW snapshot_validator as
select *,
       rank() over (partition by snapshot_at_epoch_number order by total_stake_amount_nanos desc, badger_key desc) -
       1 as validator_rank
from snapshot_validator_entry;

CREATE OR REPLACE VIEW epoch_details_for_block as
select block_hash, epoch_number, bls.pkid as proposer_pkid, epoch_entry.snapshot_at_epoch_number, height
from block
         left join epoch_entry
                   on block.height between epoch_entry.initial_block_height and epoch_entry.final_block_height
         left join bls_public_key_pkid_pair_snapshot_entry bls
                   on bls.snapshot_at_epoch_number = epoch_entry.snapshot_at_epoch_number and
                      block.proposer_voting_public_key = bls.bls_public_key;
CREATE UNIQUE INDEX epoch_details_for_block_unique_index ON epoch_details_for_block (block_hash, epoch_number);

CREATE OR REPLACE VIEW block_validator_signers as
select sv.*, edfb.block_hash, edfb.height
from epoch_details_for_block edfb
         join snapshot_validator sv on edfb.snapshot_at_epoch_number = sv.snapshot_at_epoch_number
         join block_signer bs on bs.block_hash = edfb.block_hash and bs.signer_index = sv.validator_rank-1;

comment on table snapshot_validator_entry is E'@omit';

comment on column snapshot_validator.badger_key is E'@omit';

comment on view snapshot_validator is E'@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName snapshotValidatorEntries|@fieldName validatorEntry';
comment on view block_validator_signers is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName snapshotValidatorSigners|@fieldName block\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName signedBlocks|@fieldName validatorEntry';

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
       coalesce(total_rewards, 0)                                                                                                                                                                                            as total_stake_reward_nanos,
       latest_block_signed.height                                                                                                                                                                                            as latest_block_signed
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
                   on total_stake_rewards.validator_pkid = validator_entry.validator_pkid
         left join (select validator_pkid, max(height) as height
                    from block_validator_signers bvs
                    group by validator_pkid) as latest_block_signed
                   on latest_block_signed.validator_pkid = validator_entry.validator_pkid;

CREATE UNIQUE INDEX validator_stats_unique_index ON validator_stats (validator_pkid);
comment on materialized view validator_stats is E'@primaryKey validator_pkid\n@unique validator_rank\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName validatorStats|@fieldName validatorEntry';
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
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
drop view if exists block_validator_signers CASCADE;
DROP VIEW IF EXISTS snapshot_validator CASCADE;
CREATE OR REPLACE VIEW epoch_details_for_block as
select block_hash, epoch_number, bls.pkid as proposer_pkid
from block
         left join epoch_entry
                   on epoch_entry.initial_block_height <= block.height and
                      epoch_entry.final_block_height >= block.height
         left join bls_public_key_pkid_pair_snapshot_entry bls
                   on bls.snapshot_at_epoch_number = epoch_entry.snapshot_at_epoch_number and
                      block.proposer_voting_public_key = bls.bls_public_key;
comment on table snapshot_validator_entry is null;

comment on column snapshot_validator.badger_key is null;
`)
		if err != nil {
			return err
		}

		return nil
	})
}
