package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		// Create a new table to handle web push notification subscriptions.
		_, err := db.Exec(`
			CREATE TABLE post_entry (
				post_hash VARCHAR PRIMARY KEY,
				poster_public_key VARCHAR,
				parent_post_hash VARCHAR,
				body TEXT,
				image_urls VARCHAR[],
				video_urls VARCHAR[],
				reposted_post_hash VARCHAR,
				quoted_repost BOOLEAN,
				timestamp TIMESTAMP,
				hidden BOOLEAN,
				like_count BIGINT,
				repost_count BIGINT,
				quote_repost_count BIGINT,
				diamond_count BIGINT,
				comment_count BIGINT,
				pinned BOOLEAN,
				nft BOOLEAN,
				num_nft_copies BIGINT,
				num_nft_copies_for_sale BIGINT,
				num_nft_copies_burned BIGINT,
				unlockable BOOLEAN,
				creator_royalty_basis_points BIGINT,
				coin_royalty_basis_points BIGINT,
				additional_nft_royalties_to_coins_basis_points JSONB,
				additional_nft_royalties_to_creators_basis_points JSONB,
				extra_data JSONB,
				is_frozen BOOLEAN,
				badger_key bytea
			);
			CREATE INDEX badger_key_idx ON post_entry (badger_key);
			CREATE INDEX timestamp_idx ON post_entry (timestamp desc);
			CREATE INDEX nft_idx ON post_entry (nft);
			-- NOTE: It would be nice for these to be foreign keys, but that would require consensus to
			-- be able to handle the case where a post is deleted, which is not currently done. 
			CREATE INDEX reposted_post_hash_idx ON post_entry (reposted_post_hash);
			CREATE INDEX parent_post_hash_idx ON post_entry (parent_post_hash);
			CREATE INDEX poster_public_key_timestamp_idx ON post_entry (poster_public_key, timestamp DESC);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE post_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
