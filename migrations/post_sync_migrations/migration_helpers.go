package post_sync_migrations

import (
	"errors"
	"fmt"
	"github.com/uptrace/bun"
	"math"
	"time"
)

const (
	retryLimit = 5
)

func RunMigrationWithRetries(db *bun.DB, migrationQuery string) error {
	for i := 0; i < retryLimit; i++ {
		_, err := db.Exec(migrationQuery)
		if err == nil {
			return nil
		}
		waitTime := 5 * time.Duration(math.Pow(2, float64(i))) * time.Second
		fmt.Printf("Failed to migrate, retrying in %v: %v\n", waitTime, err)
		time.Sleep(waitTime)
	}
	return errors.New("Failed to migrate after 5 attempts")
}
