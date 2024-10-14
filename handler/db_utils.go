package handler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/deso-protocol/postgres-data-handler/migrations/initial_migrations"
	"github.com/deso-protocol/postgres-data-handler/migrations/post_sync_migrations"
	"github.com/golang/glog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

type MigrationType uint8

const (
	MigrationContextKey = "migration_context"
)

func RunMigrations(db *bun.DB, migrations *migrate.Migrations, ctx context.Context) error {
	var migrator *migrate.Migrator

	// Make sure we don't mark a migration as successful if it fails.
	migrationOpt := migrate.WithMarkAppliedOnSuccess(true)
	migrator = migrate.NewMigrator(db, migrations, migrationOpt)

	if err := AcquireAdvisoryLock(db); err != nil {
		return err
	}
	defer func() {
		if err := ReleaseAdvisoryLock(db); err != nil {
			glog.Errorf("Error releasing advisory lock: %v", err)
		}
	}()
	if err := migrator.Init(ctx); err != nil {
		glog.Fatal(err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	glog.Infof("Migrated to %s\n", group)
	return nil
}

func RollbackAllMigrations(migrator *migrate.Migrator, ctx context.Context) error {
	// Get all applied migrations
	appliedMigrations, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Loop through every applied migration, rolling back each one.
	for _, _ = range appliedMigrations {
		if _, err = migrator.Rollback(ctx); err != nil {
			return err
		}
	}
	return nil
}

type DBConfig struct {
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBName     string
}

func SetupDb(dbConfig *DBConfig, threadLimit int, logQueries bool, readonlyUserPassword string, calculateExplorerStatistics bool) (*bun.DB, error) {
	pgURI := PGUriFromDbConfig(dbConfig)
	// Open a PostgreSQL database.
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(pgURI)))
	if pgdb == nil {
		glog.Fatalf("Error connecting to postgres db at URI: %v", pgURI)
	}

	// Create a Bun db on top of postgres for querying.
	db := bun.NewDB(pgdb, pgdialect.New())

	db.SetConnMaxLifetime(0)

	db.SetMaxIdleConns(threadLimit * 2)

	//Print all queries to stdout for debugging.
	if logQueries {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Set the readonly user password for the initial migrations.
	initial_migrations.SetQueryUserPassword(readonlyUserPassword)

	post_sync_migrations.SetCalculateExplorerStatistics(calculateExplorerStatistics)

	ctx := CreateMigrationContext(context.Background(), dbConfig)
	// Apply db migrations.
	err := RunMigrations(db, initial_migrations.Migrations, ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateMigrationContext(ctx context.Context, config *DBConfig) context.Context {
	if config != nil {
		ctx = context.WithValue(ctx, MigrationContextKey, config)
	}
	return ctx
}

func PGUriFromDbConfig(config *DBConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timeout=18000s", config.DBUsername, config.DBPassword, config.DBHost, config.DBPort, config.DBName)
}
