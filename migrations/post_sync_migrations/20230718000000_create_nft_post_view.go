package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE VIEW nft_post AS
			SELECT
				nft_entry.owner_pkid as nft_owner_pkid,
				post_entry.post_hash,
				post_entry.timestamp
			FROM nft_entry
			JOIN post_entry on post_entry.post_hash = nft_entry.nft_post_hash
			GROUP BY nft_entry.nft_post_hash, nft_entry.owner_pkid, post_entry.post_hash;

			COMMENT ON VIEW nft_post IS E'@foreignKey (nft_owner_pkid) references account (pkid)|@foreignFieldName nftPostsAsOwner|@fieldName owner\n@foreignKey (post_hash) references post_entry (post_hash)|@foreignFieldName nftPosts|@fieldName post';
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			COMMENT ON VIEW nft_post IS NULL;
			DROP VIEW nft_post;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
