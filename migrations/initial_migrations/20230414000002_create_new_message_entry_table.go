package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE new_message_entry (
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
		
		CREATE INDEX new_message_sender_access_group_owner_public_key_idx ON new_message_entry (sender_access_group_owner_public_key);
		CREATE INDEX new_message_sender_access_group_key_name_idx ON new_message_entry (sender_access_group_key_name);
		CREATE INDEX new_message_sender_access_group_public_key_idx ON new_message_entry (sender_access_group_public_key);
		CREATE INDEX new_message_recipient_access_group_owner_public_key_idx ON new_message_entry (recipient_access_group_owner_public_key);
		CREATE INDEX new_message_recipient_access_group_key_name_idx ON new_message_entry (recipient_access_group_key_name);
		CREATE INDEX new_message_recipient_access_group_public_key_idx ON new_message_entry (recipient_access_group_public_key);
		CREATE INDEX new_message_sender_access_group_public_key_timestamp_idx ON new_message_entry (sender_access_group_public_key, timestamp desc);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE new_message_entry
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
