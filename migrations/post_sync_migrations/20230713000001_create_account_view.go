package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE VIEW account AS
			SELECT
				wallet.pkid,
				wallet.public_key,
				profile_entry.username,
				profile_entry.description,
				profile_entry.profile_pic,
				profile_entry.creator_basis_points,
				profile_entry.coin_watermark_nanos,
				profile_entry.minting_disabled,
				profile_entry.dao_coin_minting_disabled,
				profile_entry.dao_coin_transfer_restriction_status,
				profile_entry.extra_data,
				profile_entry.coin_price_deso_nanos,
				profile_entry.deso_locked_nanos,
				profile_entry.cc_coins_in_circulation_nanos,
        		profile_entry.dao_coins_in_circulation_nanos_hex,
				true as token_balance_join_field,
                false as cc_balance_join_field
			FROM
				wallet
			LEFT JOIN
				profile_entry
			ON
				wallet.public_key = profile_entry.public_key;
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS account CASCADE;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
