package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createDesoBalanceEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				pkid            VARCHAR NOT NULL,
				balance_nanos   BIGINT NOT NULL,
				badger_key      BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_pkid_idx ON {tableName} (pkid);
			-- TODO: Define FK relations
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		if err := createDesoBalanceEntryTable(db, "deso_balance_entry"); err != nil {
			return err
		}
		if err := createDesoBalanceEntryTable(db, "deso_balance_entry_utxo_ops"); err != nil {
			return err
		}
		return AddUtxoOpColumnsToTable(db, "deso_balance_entry_utxo_ops")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE deso_balance_entry;
			DROP TABLE deso_balance_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
