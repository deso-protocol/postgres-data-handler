package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		// Make sure work_mem is set to a sufficient amount
		_, err := db.Exec(`
			SET work_mem = '32MB';
		`)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			SET work_mem = '4MB';
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
