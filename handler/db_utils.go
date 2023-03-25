package handler

import (
	"PostgresDataHandler/migrations/initial_migrations"
	"context"
	"github.com/golang/glog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

type MigrationType uint8

const (
	// We intentionally skip zero as otherwise that would be the default value.
	MigrationTypeInitial       MigrationType = 0
	MigrationTypePostHypersync MigrationType = 1
)

func RunMigrations(db *bun.DB, reset bool, migrationType MigrationType) error {
	ctx := context.Background()
	var migrator *migrate.Migrator

	if migrationType == MigrationTypeInitial {
		migrator = migrate.NewMigrator(db, initial_migrations.Migrations)
	} else {
		// TODO: Determine whether we need to add migrations for post-hypersync.
	}
	if err := migrator.Init(ctx); err != nil {
		glog.Fatal(err)
	}

	if reset {
		if err := RollbackAllMigrations(migrator, ctx); err != nil {
			return err
		}
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
