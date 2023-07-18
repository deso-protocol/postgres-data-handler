package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createDiamondEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				sender_pkid            VARCHAR NOT NULL,
				receiver_pkid          VARCHAR NOT NULL,
				post_hash              VARCHAR NOT NULL,
				diamond_level          SMALLINT NOT NULL,
				badger_key             BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_sender_public_key_idx ON {tableName} (sender_pkid);
			CREATE INDEX {tableName}_receiver_public_key_idx ON {tableName} (receiver_pkid);
			CREATE INDEX {tableName}_post_hash_idx ON {tableName} (post_hash);
			-- TODO: Define FK relations
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createDiamondEntryTable(db, "diamond_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE diamond_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
