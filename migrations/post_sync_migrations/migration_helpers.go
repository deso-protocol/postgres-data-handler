package post_sync_migrations

import (
	"context"
	"errors"
	"fmt"
	"github.com/uptrace/bun"
	"math"
	"time"
)

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
		fmt.Printf("Failed to migrate, retrying in %v: %v\n", waitTime, err)
		time.Sleep(waitTime)
	}
	return errors.New("Failed to migrate after 5 attempts")
}
