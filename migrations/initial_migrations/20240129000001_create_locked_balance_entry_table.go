package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createLockedBalanceEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				hodler_pkid VARCHAR NOT NULL,
				profile_pkid VARCHAR NOT NULL,
				unlock_timestamp_nano_secs BIGINT NOT NULL,
				vesting_end_timestamp_nano_secs BIGINT NOT NULL,
				balance_base_units NUMERIC(78, 0) NOT NULL,

				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX {tableName}_hodler_pkid_idx ON {tableName} (hodler_pkid);
			CREATE INDEX {tableName}_profile_pkid_idx ON {tableName} (profile_pkid);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createLockedBalanceEntryTable(db, "locked_balance_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS locked_balance_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
