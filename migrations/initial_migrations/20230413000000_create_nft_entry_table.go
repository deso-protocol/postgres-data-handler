package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createNftEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				last_owner_pkid VARCHAR,
				owner_pkid VARCHAR NOT NULL,
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
			CREATE INDEX {tableName}_owner_pkid_idx ON {tableName} (owner_pkid);
			CREATE INDEX {tableName}_nft_post_hash_idx ON {tableName} (nft_post_hash);
			CREATE INDEX {tableName}_serial_number_idx ON {tableName} (serial_number);
			CREATE INDEX {tableName}_is_for_sale_idx ON {tableName} (is_for_sale);
			CREATE INDEX {tableName}_is_buy_now_idx ON {tableName} (is_buy_now);
			CREATE INDEX {tableName}_min_bid_amount_nanos_idx ON {tableName} (min_bid_amount_nanos desc);
			CREATE INDEX {tableName}_is_for_sale_min_bid_amount_nanos_idx ON {tableName} (is_for_sale, min_bid_amount_nanos desc);
			CREATE INDEX {tableName}_is_buy_now_buy_now_price_nanos_idx ON {tableName} (is_for_sale, buy_now_price_nanos desc);
			CREATE INDEX {tableName}_buy_now_price_nanos_idx ON {tableName} (buy_now_price_nanos desc);
			CREATE INDEX {tableName}_is_pending_idx ON {tableName} (is_pending);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createNftEntryTable(db, "nft_entry")
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
