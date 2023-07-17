package post_sync_migrations

import (
	"PostgresDataHandler/migrations/initial_migrations"
	"context"
	"github.com/uptrace/bun"
)

func init() {
	initial_migrations.Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
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
				profile_entry.extra_data
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
			DROP VIEW account;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
