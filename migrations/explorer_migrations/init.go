package explorer_migrations

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"math"
	"time"
)

var (
	Migrations                  = migrate.NewMigrations()
)

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}

const (
	retryLimit = 10
)

func RunMigrationWithRetries(db *bun.DB, migrationQuery string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()
	for ii := 0; ii < retryLimit; ii++ {
		_, err := db.ExecContext(ctx, migrationQuery)
		if err == nil {
			return nil
		}
		waitTime := 5 * time.Duration(math.Pow(2, float64(ii))) * time.Second
		fmt.Printf("Failed to migrate, retrying in %v. err: %v. Query: %v\n", waitTime, err, migrationQuery)
		time.Sleep(waitTime)
	}
	return errors.New("Failed to migrate after 5 attempts")
}
