package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
                comment on column global_params_entry.badger_key is E'@omit';
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
                comment on column global_params_entry.badger_key is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
