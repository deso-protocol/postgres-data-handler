package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE access_group_entry (
				access_group_owner_public_key VARCHAR,
				access_group_key_name VARCHAR,
				access_group_public_key VARCHAR,
			
				extra_data jsonb,
				badger_key BYTEA PRIMARY KEY
		);
		
		CREATE INDEX access_group_public_key_idx ON access_group_entry (access_group_public_key);
		CREATE INDEX access_group_key_name_idx ON access_group_entry (access_group_key_name);
		CREATE INDEX access_group_owner_public_key_idx ON access_group_entry (access_group_owner_public_key);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE access_group_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
