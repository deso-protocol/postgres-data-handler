package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createAccessGroupMemberEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				access_group_member_public_key VARCHAR NOT NULL DEFAULT '',
				access_group_owner_public_key VARCHAR NOT NULL DEFAULT '',
				access_group_member_key_name VARCHAR NOT NULL DEFAULT '',
				access_group_key_name VARCHAR NOT NULL DEFAULT '',
				encrypted_key BYTEA NOT NULL DEFAULT '',
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY
			);
		
		CREATE INDEX {tableName}_public_key_idx ON {tableName} (access_group_member_public_key);
		CREATE INDEX {tableName}_member_key_name_idx ON {tableName} (access_group_member_key_name);
		CREATE INDEX {tableName}_key_name_idx ON {tableName} (access_group_key_name);
		CREATE INDEX {tableName}_owner_public_key_idx ON {tableName} (access_group_owner_public_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createAccessGroupMemberEntryTable(db, "access_group_member_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS access_group_member_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
