package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE balance_entry (
			    hodler_pkid					 VARCHAR NOT NULL,
			    creator_pkid				 VARCHAR NOT NULL,
				balance_nanos	             NUMERIC(78, 0) NOT NULL,
				has_purchased	             BOOLEAN NOT NULL,
				is_dao_coin		             BOOLEAN NOT NULL,
				badger_key                   BYTEA PRIMARY KEY
			);
			CREATE INDEX balance_hodler_pkid_idx ON balance_entry (hodler_pkid);
			CREATE INDEX balance_creator_pkid_idx ON balance_entry (creator_pkid);
			CREATE INDEX balance_has_purchased_idx ON balance_entry (has_purchased);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE balance_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
