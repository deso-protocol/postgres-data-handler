package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createFollowEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				follower_pkid          VARCHAR NOT NULL,
				followed_pkid          VARCHAR NOT NULL,
				badger_key             BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_follower_idx ON {tableName} (follower_pkid);
			CREATE INDEX {tableName}_followed_idx ON {tableName} (followed_pkid);
			CREATE UNIQUE INDEX {tableName}_badger_key_idx ON {tableName} (badger_key);
			-- TODO: Define FK relations
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createFollowEntryTable(db, "follow_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE follow_entry;
			DROP TABLE follow_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
