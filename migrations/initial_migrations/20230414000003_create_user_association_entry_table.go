package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createUserAssociationEntry(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				association_id VARCHAR,
				transactor_pkid VARCHAR,
				target_user_pkid VARCHAR,
				app_pkid VARCHAR,
				association_type VARCHAR NOT NULL,
				association_value VARCHAR NOT NULL,
				block_height INTEGER,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY
			);
		
		CREATE INDEX {tableName}_association_id_idx ON {tableName} (association_id);
		CREATE INDEX {tableName}_transactor_pkid_idx ON {tableName} (transactor_pkid);
		CREATE INDEX {tableName}_target_user_pkid_idx ON {tableName} (target_user_pkid);
		CREATE INDEX {tableName}_app_pkid_idx ON {tableName} (app_pkid);
		CREATE INDEX {tableName}_association_type_idx ON {tableName} (association_type);
		CREATE INDEX {tableName}_block_height_idx ON {tableName} (block_height desc);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createUserAssociationEntry(db, "user_association_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS user_association_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
