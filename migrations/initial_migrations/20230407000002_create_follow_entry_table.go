package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE follow_entry (
				follower_pkid          BYTEA NOT NULL,
				followed_pkid          BYTEA NOT NULL,
				badger_key             BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX follow_follower_idx ON follow_entry (follower_pkid);
			CREATE INDEX follow_followed_idx ON follow_entry (followed_pkid);
			CREATE UNIQUE INDEX follow_badger_key_idx ON follow_entry (badger_key);
			-- TODO: Define FK relations
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE follow_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
