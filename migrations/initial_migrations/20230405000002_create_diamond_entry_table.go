package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE diamond_entry (
				sender_pkid            BYTEA NOT NULL,
				receiver_pkid          BYTEA NOT NULL,
				post_hash              VARCHAR NOT NULL,
				diamond_level          SMALLINT NOT NULL,
				badger_key             BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX diamond_sender_public_key_idx ON diamond_entry (sender_pkid);
			CREATE INDEX diamond_receiver_public_key_idx ON diamond_entry (receiver_pkid);
			CREATE INDEX diamond_post_hash_idx ON diamond_entry (post_hash);
			CREATE UNIQUE INDEX diamond_badger_key_idx ON diamond_entry (badger_key);
			-- TODO: Define FK relations
		`)
		if err != nil {
			return err
		}
		return nil
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
