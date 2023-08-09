package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE VIEW creator_coin_balance AS
			SELECT
				balance_entry.hodler_pkid,
				balance_entry.creator_pkid,
				balance_entry.balance_nanos,
				balance_entry.has_purchased,
    		profile_entry.coin_price_deso_nanos,
    		(balance_entry.balance_nanos * coin_price_deso_nanos)::NUMERIC AS total_value_nanos
			FROM balance_entry
			JOIN profile_entry ON profile_entry.pkid = balance_entry.creator_pkid
			WHERE is_dao_coin = false;
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS creator_coin_balance;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
