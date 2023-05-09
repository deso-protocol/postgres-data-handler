package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE pkid_entry (
				pkid            VARCHAR NOT NULL,
				public_key      VARCHAR NOT NULL,
				badger_key      BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX pkid_entry_pkid_idx ON pkid_entry (pkid);
			CREATE INDEX pkid_entry_public_key ON pkid_entry (public_key);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE pkid_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
