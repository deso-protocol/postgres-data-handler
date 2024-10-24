package explorer_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		create index if not exists statistic_profile_transactions_latest_idx on statistic_profile_transactions (latest_transaction_timestamp desc);
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		drop index if exists statistic_profile_transactions_latest_idx;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
