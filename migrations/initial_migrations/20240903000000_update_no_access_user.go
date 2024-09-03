package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create noaccess role in db.
		_, err := db.Exec(`
			REVOKE ALL ON ALL TABLES IN SCHEMA public FROM noaccess;
			REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM noaccess;
			REVOKE ALL ON ALL FUNCTIONS IN SCHEMA public FROM noaccess;
			REVOKE ALL ON SCHEMA public FROM noaccess;
			GRANT USAGE ON SCHEMA public TO noaccess;
			GRANT SELECT ON ALL TABLES IN SCHEMA pg_catalog TO noaccess;
			REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM noaccess;
			grant noaccess to query_user;	
		`)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		// Lastly, delete the noaccess role.
		_, err := db.Exec(`
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM noaccess;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM noaccess;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA public FROM noaccess;
REVOKE ALL ON SCHEMA public FROM noaccess;
REVOKE USAGE ON SCHEMA public TO noaccess;
REVOKE SELECT ON ALL TABLES IN SCHEMA pg_catalog TO noaccess;
REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM noaccess;
grant noaccess to query_user;
`)
		if err != nil {
			return err
		}

		return nil
	})
}
