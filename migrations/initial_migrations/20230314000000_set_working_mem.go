package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		// Make sure work_mem is set to a sufficient amount
		_, err := db.Exec(`
			CREATE EXTENSION IF NOT EXISTS pg_trgm;

		`)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP EXTENSION IF EXISTS pg_trgm;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
