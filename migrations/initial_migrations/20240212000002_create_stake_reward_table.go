package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createStakeRewardTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				staker_pkid VARCHAR NOT NULL,
				validator_pkid VARCHAR NOT NULL,
				reward_method SMALLINT NOT NULL,
				reward_nanos BIGINT NOT NULL,
				is_validator_commission BOOLEAN NOT NULL,
				block_hash VARCHAR NOT NULL,
				utxo_op_index BIGINT NOT NULL,
				PRIMARY KEY(block_hash, utxo_op_index)
			);
			CREATE INDEX {tableName}_validator_pkid_idx ON {tableName} (validator_pkid);
			CREATE INDEX {tableName}_staker_pkid_idx ON {tableName} (staker_pkid);
			CREATE INDEX {tableName}_block_hash_idx ON {tableName} (block_hash);
			CREATE INDEX {tableName}_is_validator_commission_idx ON {tableName} (is_validator_commission);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createStakeRewardTable(db, "stake_reward")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS stake_reward;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
