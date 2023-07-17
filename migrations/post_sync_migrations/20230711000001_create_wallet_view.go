package post_sync_migrations

import (
	"PostgresDataHandler/migrations/initial_migrations"
	"context"
	"github.com/uptrace/bun"
)

func init() {
	initial_migrations.Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE VIEW wallet AS
			SELECT pkid, public_key FROM pkid_entry
			UNION ALL
			SELECT public_key AS pkid, public_key
			FROM public_key
			WHERE public_key NOT IN (SELECT public_key FROM pkid_entry);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP VIEW wallet;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
