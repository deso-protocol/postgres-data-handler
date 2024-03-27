package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createYieldCurvePointTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				profile_pkid VARCHAR NOT NULL,
				lockup_duration_nano_secs BIGINT NOT NULL,
				lockup_yield_apy_basis_points BIGINT NOT NULL,

				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX {tableName}_profile_pkid_idx ON {tableName} (profile_pkid);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createYieldCurvePointTable(db, "yield_curve_point")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS yield_curve_point;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
