package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createGlobalParamsEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				usd_cents_per_bitcoin BIGINT NOT NULL,
				create_profile_fee_nanos BIGINT NOT NULL,
				create_nft_fee_nanos BIGINT NOT NULL,
				max_copies_per_nft BIGINT NOT NULL,
				minimum_network_fee_nanos_per_kb BIGINT NOT NULL,
				max_nonce_expiration_block_height_offset BIGINT NOT NULL,
				stake_lockup_epoch_duration BIGINT NOT NULL,
				validator_jail_epoch_duration BIGINT NOT NULL,
				leader_schedule_max_num_validators BIGINT NOT NULL,
				validator_set_max_num_validators BIGINT NOT NULL,
				staking_rewards_max_num_stakes BIGINT NOT NULL,
				staking_rewards_apy_basis_points BIGINT NOT NULL,
				epoch_duration_num_blocks BIGINT NOT NULL,
				jail_inactive_validator_grace_period_epochs BIGINT NOT NULL,
				maximum_vested_intersections_per_lockup_transaction INT NOT NULL,
				fee_bucket_growth_rate_basis_points BIGINT NOT NULL,
				block_timestamp_drift_nano_secs BIGINT NOT NULL,
				mempool_max_size_bytes BIGINT NOT NULL,
				mempool_fee_estimator_num_mempool_blocks BIGINT NOT NULL,
				mempool_fee_estimator_num_past_blocks BIGINT NOT NULL,
				max_block_size_bytes_pos BIGINT NOT NULL,
				soft_max_block_size_bytes_pos BIGINT NOT NULL,
				max_txn_size_bytes_pos BIGINT NOT NULL,
				block_production_interval_milliseconds_pos BIGINT NOT NULL,
				timeout_interval_milliseconds_pos BIGINT NOT NULL,
				badger_key BYTEA PRIMARY KEY 
			);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createGlobalParamsEntryTable(db, "global_params_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS global_params_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
