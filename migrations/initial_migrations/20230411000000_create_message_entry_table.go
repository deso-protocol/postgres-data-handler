package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE message_entry (
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
			CREATE INDEX message_sender_public_key_idx ON message_entry (sender_public_key);
			CREATE INDEX message_recipient_public_key_idx ON message_entry (recipient_public_key);
			CREATE INDEX message_recipient_timestamp_public_key_idx ON message_entry (recipient_public_key, timestamp desc);
			CREATE INDEX message_version_idx ON message_entry (version);
			CREATE INDEX message_sender_messaging_public_key_idx ON message_entry (sender_messaging_public_key);
			CREATE INDEX message_recipient_messaging_public_key_idx ON message_entry (recipient_messaging_public_key);
			CREATE INDEX message_sender_messaging_group_key_name_idx ON message_entry (sender_messaging_group_key_name);
			CREATE INDEX message_recipient_messaging_group_key_name_idx ON message_entry (recipient_messaging_group_key_name);
			CREATE INDEX message_recipient_messaging_group_key_name_timestamp_idx ON message_entry (recipient_messaging_group_key_name, timestamp desc);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE message_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
