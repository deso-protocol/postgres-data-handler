package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE deso_balance_entry (
				pkid            VARCHAR NOT NULL,
				balance_nanos   BIGINT NOT NULL,
				badger_key      BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX deso_balance_entry_pkid_idx ON deso_balance_entry (pkid);
			-- TODO: Define FK relations
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE deso_balance_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
