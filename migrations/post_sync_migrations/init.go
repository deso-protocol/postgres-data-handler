package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"time"
)

var (
	calculateExplorerStatistics bool
	Migrations                  = migrate.NewMigrations()
)

func SetCalculateExplorerStatistics(calculate bool) {
	calculateExplorerStatistics = calculate
}

func executeQuery(db *bun.DB, query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	_, err := db.Exec(query, ctx)
	return err
}

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}
