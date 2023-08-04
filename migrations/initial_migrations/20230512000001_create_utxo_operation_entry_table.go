package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
	"strings"
)

func AddUtxoOpColumnsToTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			ALTER TABLE {tableName}
			DROP CONSTRAINT IF EXISTS {tableName}_pkey;
			ALTER TABLE {tableName}
			ALTER COLUMN badger_key DROP NOT NULL;
			ALTER TABLE {tableName}
			ADD COLUMN utxo_op_entry_type VARCHAR;
			ALTER TABLE {tableName}
			ADD COLUMN utxo_op_index INTEGER;
			ALTER TABLE {tableName}
			ADD COLUMN transaction_index INTEGER;
			ALTER TABLE {tableName}
			ADD COLUMN array_index INTEGER;
			ALTER TABLE {tableName}
			ADD COLUMN block_hash VARCHAR;
			CREATE INDEX {tableName}_utxo_op_entry_type_utxo_op_index_transaction_index_block_hash_idx on {tableName} (utxo_op_entry_type, utxo_op_index, transaction_index, block_hash);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE utxo_operation (
				operation_type            INTEGER NOT NULL,
				block_hash                VARCHAR NOT NULL,
				transaction_index		  INTEGER NOT NULL,
				utxo_op_index             INTEGER NOT NULL,
				utxo_op_bytes			  BYTEA NOT NULL
				badger_key				  BYTEA NOT NULL,
			);
			CREATE UNIQUE INDEX utxo_operation_block_hash_transaction_utxo_op_idx ON utxo_operation (block_hash, transaction_index desc, utxo_op_index desc);
			CREATE INDEX utxo_entry_type_block_hash_transaction_utxo_op_idx ON utxo_operation (operation_type, block_hash, transaction_index desc, utxo_op_index desc);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS utxo_operation;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
