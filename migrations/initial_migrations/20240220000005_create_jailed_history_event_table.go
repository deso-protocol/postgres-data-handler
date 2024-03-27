package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createJailedHistoryEventTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				validator_pkid VARCHAR NOT NULL,
				jailed_at_epoch_number BIGINT NOT NULL,
				unjailed_at_epoch_number BIGINT NOT NULL,
				PRIMARY KEY(validator_pkid, jailed_at_epoch_number, unjailed_at_epoch_number)	
			);
			CREATE INDEX {tableName}_validator_pkid_idx ON {tableName} (validator_pkid);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createJailedHistoryEventTable(db, "jailed_history_event")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS jailed_history_event;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
