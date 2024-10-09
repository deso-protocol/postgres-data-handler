package explorer_view_migrations

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
)

const (
	MigrationContextKey = "migration_context"
)

type DBConfig struct {
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBName     string
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		dbConfig, ok := ctx.Value(MigrationContextKey).(*DBConfig)
		if !ok {
			return fmt.Errorf("could not get config from context")
		}
		userPassword := dbConfig.DBPassword
		userName := dbConfig.DBUsername
		dbName := dbConfig.DBName
		host := dbConfig.DBHost
		port := dbConfig.DBPort

		if _, err := db.Exec(`
			CREATE EXTENSION IF NOT EXISTS postgres_fdw;
			
			-- Create a foreign server
			CREATE SERVER IF NOT EXISTS subscriber_server
			FOREIGN DATA WRAPPER postgres_fdw
			OPTIONS (host ?, port ?, dbname ?);
			
			-- Create a user mapping
			CREATE USER MAPPING IF NOT EXISTS FOR current_user
			SERVER subscriber_server
			OPTIONS (user ?, password ?);

			IMPORT FOREIGN SCHEMA public
			LIMIT TO (public_key_first_transaction)
			FROM SERVER subscriber_server
			INTO public;
		`, host, port, dbName, userName, userPassword); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		drop index if exists statistic_profile_transactions_latest_idx;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
