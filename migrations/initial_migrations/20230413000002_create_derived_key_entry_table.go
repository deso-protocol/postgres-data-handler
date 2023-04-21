package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE derived_key_entry (
				owner_public_key VARCHAR NOT NULL,
				derived_public_key VARCHAR NOT NULL,
				expiration_block BIGINT NOT NULL,
				operation_type SMALLINT NOT NULL,
			
				global_deso_limit BIGINT,
				is_unlimited BOOLEAN,
				transaction_spending_limit_bytes BYTEA,
			
				extra_data jsonb,
				badger_key BYTEA PRIMARY KEY
			);

			CREATE INDEX derived_key_owner_public_key_idx ON derived_key_entry (owner_public_key);
			CREATE INDEX derived_key_derived_public_key_idx ON derived_key_entry (derived_public_key);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE derived_key_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
