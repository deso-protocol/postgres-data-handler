package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createPkidEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				pkid            VARCHAR NOT NULL,
				public_key      VARCHAR NOT NULL,
				badger_key      BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_pkid_idx ON {tableName} (pkid);
			CREATE INDEX {tableName}_public_key ON {tableName} (public_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		if err := createPkidEntryTable(db, "pkid_entry"); err != nil {
			return err
		}
		if err := createPkidEntryTable(db, "pkid_entry_utxo_ops"); err != nil {
			return err
		}
		return AddUtxoOpColumnsToTable(db, "pkid_entry_utxo_ops")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE pkid_entry;
			DROP TABLE pkid_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
