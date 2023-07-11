package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createNftBidEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
			bidder_pkid VARCHAR NOT NULL,
			nft_post_hash VARCHAR NOT NULL,
			serial_number BIGINT NOT NULL,
			bid_amount_nanos BIGINT NOT NULL,
			accepted_block_height BIGINT,
			badger_key BYTEA PRIMARY KEY
		);
		
		CREATE INDEX {tableName}_bidder_pkid_idx ON {tableName} (bidder_pkid);
		CREATE INDEX {tableName}_nft_post_hash_idx ON {tableName} (nft_post_hash);
		CREATE INDEX {tableName}_serial_number_idx ON {tableName} (serial_number);
		CREATE INDEX {tableName}_serial_number_bid_amount_nanos_idx ON {tableName} (serial_number, bid_amount_nanos desc);
		CREATE INDEX {tableName}_bid_amount_nanos_idx ON {tableName} (bid_amount_nanos desc);
		CREATE INDEX {tableName}_accepted_block_height_idx ON {tableName} (accepted_block_height);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createNftBidEntryTable(db, "nft_bid_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE nft_bid_entry;
			DROP TABLE nft_bid_entry_utxo_ops;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
