package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE nft_bid_entry (
			bidder_pkid BYTEA NOT NULL,
			nft_post_hash VARCHAR NOT NULL,
			serial_number BIGINT NOT NULL,
			bid_amount_nanos BIGINT NOT NULL,
			accepted_block_height BIGINT,
			badger_key BYTEA PRIMARY KEY
		);
		
		CREATE INDEX nft_bid_entry_bidder_pkid_idx ON nft_bid_entry (bidder_pkid);
		CREATE INDEX nft_bid_entry_nft_post_hash_idx ON nft_bid_entry (nft_post_hash);
		CREATE INDEX nft_bid_entry_serial_number_idx ON nft_bid_entry (serial_number);
		CREATE INDEX nft_bid_entry_serial_number_bid_amount_nanos_idx ON nft_bid_entry (serial_number, bid_amount_nanos desc);
		CREATE INDEX nft_bid_entry_bid_amount_nanos_idx ON nft_bid_entry (bid_amount_nanos desc);
		CREATE INDEX nft_bid_entry_accepted_block_height_idx ON nft_bid_entry (accepted_block_height);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE nft_bid_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
