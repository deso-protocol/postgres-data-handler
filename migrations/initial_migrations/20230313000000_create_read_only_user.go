package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// If queryUserPassword is empty, then we don't need to do anything.
		if queryUserPassword == "" {
			return nil
		}
		// Create readonly role in db.
		_, err := db.Exec(`
			DO $$
			BEGIN
			   IF NOT EXISTS (
				  SELECT FROM pg_catalog.pg_roles 
				  WHERE  rolname = 'readaccess') THEN
			
				  CREATE ROLE readaccess;
			   END IF;
			END
			$$;
			GRANT USAGE ON SCHEMA public TO readaccess;
			GRANT SELECT ON ALL TABLES IN SCHEMA public TO readaccess;
			ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readaccess;
		`)
		if err != nil {
			return err
		}

		// Create readonly user and grant readonly role to it.
		_, err = db.Exec(`
		DO
		$do$
		BEGIN
		   IF NOT EXISTS (
			  SELECT FROM pg_catalog.pg_user 
			  WHERE  usename = 'query_user') THEN
		
			  CREATE USER query_user WITH PASSWORD ?;
              GRANT readaccess TO query_user;
		   END IF;
		END
		$do$;
	`, queryUserPassword)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		// If queryUserPassword is empty, then we don't need to do anything.
		if queryUserPassword == "" {
			return nil
		}
		// Revoke the readaccess role from query_user. Then reset the default privileges and revoke all permissions that the readaccess role has.
		_, err := db.Exec(`
		DO
		$do$
		BEGIN
		   IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'query_user') THEN
			  REVOKE readaccess FROM query_user;
		   END IF;
		   
		   IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'readaccess') THEN
			  REVOKE ALL ON SCHEMA public FROM readaccess;
			  REVOKE ALL ON ALL TABLES IN SCHEMA public FROM readaccess;
			  ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM readaccess;
		   END IF;
		END
		$do$;
	`)
		if err != nil {
			return err
		}

		// Next, delete the query_user.
		_, err = db.Exec(`
		DROP USER IF EXISTS query_user;
	`)
		if err != nil {
			return err
		}

		// Lastly, delete the readaccess role.
		_, err = db.Exec(`
		DROP ROLE IF EXISTS readaccess;
	`)
		if err != nil {
			return err
		}

		return nil
	})
}
