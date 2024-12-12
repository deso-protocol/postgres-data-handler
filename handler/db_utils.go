package handler

import (
	"context"
	"github.com/deso-protocol/postgres-data-handler/migrations/initial_migrations"
	"github.com/deso-protocol/postgres-data-handler/migrations/post_sync_migrations"
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

const (
	EntryCacheSize uint = 1000000 // 1M entries
)

// TODO: Make this a method on the PostgresDataHandler struct.
func RunMigrations(db *bun.DB, reset bool, migrationType MigrationType) error {
	ctx := context.Background()
	var migrator *migrate.Migrator

	initialMigrator := migrate.NewMigrator(db, initial_migrations.Migrations)
	postSyncMigrator := migrate.NewMigrator(db, post_sync_migrations.Migrations)

	if migrationType == MigrationTypeInitial {
		migrator = initialMigrator
	} else if migrationType == MigrationTypePostHypersync {
		migrator = postSyncMigrator
	}
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

	// If resetting, revert all migrations, starting with the most recently applied.
	if reset {
		if err := RollbackAllMigrations(postSyncMigrator, ctx); err != nil {
			return err
		}

		if err := RollbackAllMigrations(initialMigrator, ctx); err != nil {
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
	for range appliedMigrations {
		if _, err = migrator.Rollback(ctx); err != nil {
			return err
		}
	}
	return nil
}
