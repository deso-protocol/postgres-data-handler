package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE post_association_entry (
				association_id VARCHAR,
				transactor_pkid BYTEA,
				post_hash VARCHAR,
				app_pkid BYTEA,
				association_type VARCHAR NOT NULL,
				association_value VARCHAR NOT NULL,
				block_height INTEGER,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY
			);
		
		CREATE INDEX post_association_association_id_idx ON post_association_entry (association_id);
		CREATE INDEX post_association_transactor_pkid_idx ON post_association_entry (transactor_pkid);
		CREATE INDEX post_association_post_hash_idx ON post_association_entry (post_hash);
		CREATE INDEX post_association_app_pkid_idx ON post_association_entry (app_pkid);
		CREATE INDEX post_association_association_type_idx ON post_association_entry (association_type);
		CREATE INDEX post_association_block_height_idx ON post_association_entry (block_height desc);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE post_association_entry
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
