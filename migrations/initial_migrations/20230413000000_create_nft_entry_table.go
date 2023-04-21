package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE nft_entry (
				last_owner_pkid BYTEA NOT NULL,
				owner_pkid BYTEA NOT NULL,
				nft_post_hash VARCHAR NOT NULL,
				serial_number BIGINT NOT NULL,
				is_for_sale BOOLEAN NOT NULL,
				min_bid_amount_nanos BIGINT,
				unlockable_text TEXT,
				last_accepted_bid_amount_nanos BIGINT,
				is_pending BOOLEAN NOT NULL,
				is_buy_now BOOLEAN NOT NULL,
				buy_now_price_nanos BIGINT,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX nft_owner_pkid_idx ON nft_entry (owner_pkid);
			-- TODO: Make this index a FK.
			CREATE INDEX nft_nft_post_hash_idx ON nft_entry (nft_post_hash);
			CREATE INDEX nft_serial_number_idx ON nft_entry (serial_number);
			CREATE INDEX nft_is_for_sale_idx ON nft_entry (is_for_sale);
			CREATE INDEX nft_is_buy_now_idx ON nft_entry (is_buy_now);
			CREATE INDEX nft_min_bid_amount_nanos_idx ON nft_entry (min_bid_amount_nanos desc);
			CREATE INDEX nft_is_for_sale_min_bid_amount_nanos_idx ON nft_entry (is_for_sale, min_bid_amount_nanos desc);
			CREATE INDEX nft_is_buy_now_buy_now_price_nanos_idx ON nft_entry (is_for_sale, buy_now_price_nanos desc);
			CREATE INDEX nft_buy_now_price_nanos_idx ON nft_entry (buy_now_price_nanos desc);
			CREATE INDEX nft_is_pending_idx ON nft_entry (is_pending);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE nft_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
