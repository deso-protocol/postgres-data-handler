package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createMessageEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				sender_public_key VARCHAR NOT NULL,
				recipient_public_key VARCHAR NOT NULL,
				encrypted_text TEXT NOT NULL,
				timestamp TIMESTAMP,
				version SMALLINT NOT NULL,
				sender_messaging_public_key VARCHAR,
				recipient_messaging_public_key VARCHAR,
				sender_messaging_group_key_name VARCHAR,
				recipient_messaging_group_key_name VARCHAR,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY
			);
			CREATE INDEX {tableName}_sender_public_key_idx ON {tableName} (sender_public_key);
			CREATE INDEX {tableName}_recipient_public_key_idx ON {tableName} (recipient_public_key);
			CREATE INDEX {tableName}_recipient_timestamp_public_key_idx ON {tableName} (recipient_public_key, timestamp desc);
			CREATE INDEX {tableName}_version_idx ON {tableName} (version);
			CREATE INDEX {tableName}_sender_messaging_public_key_idx ON {tableName} (sender_messaging_public_key);
			CREATE INDEX {tableName}_recipient_messaging_public_key_idx ON {tableName} (recipient_messaging_public_key);
			CREATE INDEX {tableName}_sender_messaging_group_key_name_idx ON {tableName} (sender_messaging_group_key_name);
			CREATE INDEX {tableName}_recipient_messaging_group_key_name_idx ON {tableName} (recipient_messaging_group_key_name);
			CREATE INDEX {tableName}_recipient_messaging_group_key_name_timestamp_idx ON {tableName} (recipient_messaging_group_key_name, timestamp desc);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createMessageEntryTable(db, "message_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE message_entry;
			DROP TABLE message_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
