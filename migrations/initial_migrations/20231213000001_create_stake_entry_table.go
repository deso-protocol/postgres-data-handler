package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createStakeEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				staker_pkid VARCHAR NOT NULL,
				validator_pkid VARCHAR NOT NULL,
				reward_method SMALLINT NOT NULL,
				stake_amount_nanos NUMERIC(78, 0) NOT NULL,

				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX {tableName}_validator_pkid_idx ON {tableName} (validator_pkid);
			CREATE INDEX {tableName}_staker_pkid_idx ON {tableName} (staker_pkid);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createStakeEntryTable(db, "stake_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS stake_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
