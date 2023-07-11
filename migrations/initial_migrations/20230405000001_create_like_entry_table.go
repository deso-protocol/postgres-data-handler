package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createLikeEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				public_key                   VARCHAR NOT NULL,
				post_hash                    VARCHAR NOT NULL,
				badger_key                   BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_public_key_idx ON {tableName} (public_key);
			CREATE INDEX {tableName}_post_hash_idx ON {tableName} (post_hash);
`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createLikeEntryTable(db, "like_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE like_entry;
			DROP TABLE like_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
