package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createBalanceEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
			    hodler_pkid					 VARCHAR NOT NULL,
			    creator_pkid				 VARCHAR NOT NULL,
				balance_nanos	             NUMERIC(78, 0) NOT NULL,
				has_purchased	             BOOLEAN NOT NULL,
				is_dao_coin		             BOOLEAN NOT NULL,
				badger_key                   BYTEA PRIMARY KEY
			);
			CREATE INDEX {tableName}_hodler_pkid_idx ON {tableName} (hodler_pkid);
			CREATE INDEX {tableName}_creator_pkid_idx ON {tableName} (creator_pkid);
			CREATE INDEX {tableName}_has_purchased_idx ON {tableName} (has_purchased);
		`, "{tableName}", tableName, -1), nil)
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createBalanceEntryTable(db, "balance_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS balance_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
