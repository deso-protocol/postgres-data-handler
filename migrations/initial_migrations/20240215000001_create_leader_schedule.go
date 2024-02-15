package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createLeaderScheduleTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				validator_pkid VARCHAR NOT NULL,
				snapshot_at_epoch_number BIGINT NOT NULL,
				leader_index INTEGER NOT NULL,
				badger_key BYTEA PRIMARY KEY NOT NULL
			);
			CREATE INDEX {tableName}_validator_pkid_idx ON {tableName} (validator_pkid);
			CREATE INDEX {tableName}_snapshot_at_epoch_number_idx ON {tableName} (snapshot_at_epoch_number);
			CREATE INDEX {tableName}_snapshot_at_epoch_number_leader_index_idx ON {tableName} (snapshot_at_epoch_number, leader_index);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createLeaderScheduleTable(db, "leader_schedule_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS leader_schedule_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
