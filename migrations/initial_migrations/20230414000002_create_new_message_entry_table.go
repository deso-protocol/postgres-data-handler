package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createNewMessageEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				sender_access_group_owner_public_key VARCHAR,
				sender_access_group_key_name VARCHAR,
				sender_access_group_public_key VARCHAR,
				recipient_access_group_owner_public_key VARCHAR,
				recipient_access_group_key_name VARCHAR,
				recipient_access_group_public_key VARCHAR,
				encrypted_text VARCHAR NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY
			);
		
		CREATE INDEX {tableName}_sender_access_group_owner_public_key_idx ON {tableName} (sender_access_group_owner_public_key);
		CREATE INDEX {tableName}_sender_access_group_key_name_idx ON {tableName} (sender_access_group_key_name);
		CREATE INDEX {tableName}_sender_access_group_public_key_idx ON {tableName} (sender_access_group_public_key);
		CREATE INDEX {tableName}_recipient_access_group_owner_public_key_idx ON {tableName} (recipient_access_group_owner_public_key);
		CREATE INDEX {tableName}_recipient_access_group_key_name_idx ON {tableName} (recipient_access_group_key_name);
		CREATE INDEX {tableName}_recipient_access_group_public_key_idx ON {tableName} (recipient_access_group_public_key);
		CREATE INDEX {tableName}_sender_access_group_public_key_timestamp_idx ON {tableName} (sender_access_group_public_key, timestamp desc);
		`, "{tableName}", tableName, -1), nil)
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		if err := createNewMessageEntryTable(db, "new_message_entry"); err != nil {
			return err
		}
		if err := createNewMessageEntryTable(db, "new_message_entry_utxo_ops"); err != nil {
			return err
		}
		return AddUtxoOpColumnsToTable(db, "new_message_entry_utxo_ops")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE new_message_entry;
			DROP TABLE new_message_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
