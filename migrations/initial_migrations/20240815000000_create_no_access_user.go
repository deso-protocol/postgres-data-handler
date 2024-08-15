package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create noaccess role in db.
		_, err := db.Exec(`
			DO $$
			BEGIN
			   IF NOT EXISTS (
				  SELECT FROM pg_catalog.pg_roles 
				  WHERE  rolname = 'noaccess') THEN
			
				  CREATE ROLE noaccess;
			   END IF;
			END
			$$;
		`)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		// Lastly, delete the noaccess role.
		_, err := db.Exec(`DROP ROLE IF EXISTS noaccess;`)
		if err != nil {
			return err
		}

		return nil
	})
}
