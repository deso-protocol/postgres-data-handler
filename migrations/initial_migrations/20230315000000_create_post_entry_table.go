package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createPostEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				post_hash VARCHAR PRIMARY KEY,
				poster_public_key VARCHAR,
				parent_post_hash VARCHAR,
				body TEXT,
				image_urls VARCHAR[],
				video_urls VARCHAR[],
				reposted_post_hash VARCHAR,
				is_quoted_repost BOOLEAN,
				timestamp TIMESTAMP,
				is_hidden BOOLEAN,
				is_pinned BOOLEAN,
				is_nft BOOLEAN,
				num_nft_copies BIGINT,
				num_nft_copies_for_sale BIGINT,
				num_nft_copies_burned BIGINT,
				has_unlockable BOOLEAN,
				nft_royalty_to_creator_basis_points BIGINT,
				nft_royalty_to_coin_basis_points BIGINT,
				additional_nft_royalties_to_coins_basis_points JSONB,
				additional_nft_royalties_to_creators_basis_points JSONB,
				extra_data JSONB,
				is_frozen BOOLEAN,
				badger_key bytea
			);
			CREATE INDEX {tableName}_badger_key_idx ON {tableName} (badger_key);
			CREATE INDEX {tableName}_post_hash_idx ON {tableName} (post_hash);
			CREATE INDEX {tableName}_timestamp_idx ON {tableName} (timestamp desc);
			CREATE INDEX {tableName}_nft_idx ON {tableName} (is_nft);
-- 			NOTE: It would be nice for these to be foreign keys, but that would require consensus to
-- 			be able to handle the case where a post is deleted, which is not currently done. 
			CREATE INDEX {tableName}_reposted_post_hash_idx ON {tableName} (reposted_post_hash);
			CREATE INDEX {tableName}_parent_post_hash_idx ON {tableName} (parent_post_hash);
			CREATE INDEX {tableName}_poster_public_key_timestamp_idx ON {tableName} (poster_public_key, timestamp DESC);
			CREATE INDEX {tableName}_poster_public_key_nft_timestamp_idx ON {tableName} (poster_public_key, timestamp, is_nft DESC);
			CREATE INDEX {tableName}_nft_timestamp_idx ON {tableName} (timestamp, is_nft DESC);
			CREATE INDEX {tableName}_post_extra_data_node_id_idx
			ON {tableName} ((extra_data ->> 'Node'));
			CREATE INDEX {tableName}_post_extra_data_blog_slug_title_idx ON post_entry ((extra_data ->> 'BlogTitleSlug'));
			CREATE INDEX {tableName}_post_extra_data_keys_idx ON post_entry USING gin (extra_data jsonb_path_ops);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		// Make sure work_mem is set to a sufficient amount
		_, err := db.Exec(`
			SET work_mem = '32MB';
		`)
		if err != nil {
			return err
		}

		// Create post entry table
		return createPostEntryTable(db, "post_entry")
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
