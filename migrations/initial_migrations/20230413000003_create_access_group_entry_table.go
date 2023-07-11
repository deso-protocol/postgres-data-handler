package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createAccessGroupEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				access_group_owner_public_key VARCHAR,
				access_group_key_name VARCHAR,
				access_group_public_key VARCHAR,
			
				extra_data jsonb,
				badger_key BYTEA PRIMARY KEY
		);
		
		CREATE INDEX {tableName}_public_key_idx ON {tableName} (access_group_public_key);
		CREATE INDEX {tableName}_key_name_idx ON {tableName} (access_group_key_name);
		CREATE INDEX {tableName}_owner_public_key_idx ON {tableName} (access_group_owner_public_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createAccessGroupEntryTable(db, "access_group_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE access_group_entry;
			DROP TABLE access_group_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
