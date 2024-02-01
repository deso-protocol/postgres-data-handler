package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
// TODO: indexes
func createEpochEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				epoch_number BIGINT NOT NULL,
				initial_block_height BIGINT NOT NULL,
				initial_view BIGINT NOT NULL,
				final_block_height BIGINT NOT NULL,
				created_at_block_timestamp_nano_secs BIGINT NOT NULL,

				badger_key BYTEA PRIMARY KEY 
			);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createEpochEntryTable(db, "epoch_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS epoch_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
