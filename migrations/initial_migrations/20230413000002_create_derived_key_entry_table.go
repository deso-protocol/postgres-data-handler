package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createDerivedKeyEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
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

			CREATE INDEX {tableName}_owner_public_key_idx ON {tableName} (owner_public_key);
			CREATE INDEX {tableName}_derived_public_key_idx ON {tableName} (derived_public_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createDerivedKeyEntryTable(db, "derived_key_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS derived_key_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
