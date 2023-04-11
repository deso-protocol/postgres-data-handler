package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE like_entry (
				public_key                   VARCHAR NOT NULL,
				post_hash                    VARCHAR NOT NULL,
				badger_key                   BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX like_public_key_idx ON like_entry (public_key);
			CREATE INDEX like_post_hash_idx ON like_entry (post_hash);
			CREATE UNIQUE INDEX like_badger_key_idx ON like_entry (badger_key);
			-- TODO: Define FK relations
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE like_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
