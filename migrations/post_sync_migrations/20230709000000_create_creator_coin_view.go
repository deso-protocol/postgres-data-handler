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
    		(((profile_entry.deso_locked_nanos * (1 - POWER(1 - balance_entry.balance_nanos / profile_entry.cc_coins_in_circulation_nanos, 1 / 0.3333333))) * 9999) / 10000)::NUMERIC AS total_value_nanos
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
