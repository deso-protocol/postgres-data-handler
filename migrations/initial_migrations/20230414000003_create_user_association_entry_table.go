package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE user_association_entry (
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
		
		CREATE INDEX user_association_association_id_idx ON user_association_entry (association_id);
		CREATE INDEX user_association_transactor_pkid_idx ON user_association_entry (transactor_pkid);
		CREATE INDEX user_association_target_user_pkid_idx ON user_association_entry (target_user_pkid);
		CREATE INDEX user_association_app_pkid_idx ON user_association_entry (app_pkid);
		CREATE INDEX user_association_association_type_idx ON user_association_entry (association_type);
		CREATE INDEX user_association_block_height_idx ON user_association_entry (block_height desc);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE user_association_entry
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
