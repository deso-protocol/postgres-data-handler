package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create indexes for the post_entry table.
		err := RunMigrationWithRetries(db, `
 			SET work_mem = '64MB';
		`)
		if err != nil {
			return err
		}
		err = RunMigrationWithRetries(db, `
 			CREATE UNIQUE INDEX badger_key_idx ON post_entry (badger_key);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX timestamp_idx ON post_entry (timestamp desc);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX nft_idx ON post_entry (is_nft);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			-- NOTE: It would be nice for these to be foreign keys, but that would require consensus to
			-- be able to handle the case where a post is deleted, which is not currently done.
			CREATE INDEX reposted_post_hash_idx ON post_entry (reposted_post_hash);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX parent_post_hash_idx ON post_entry (parent_post_hash);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX poster_public_key_timestamp_idx ON post_entry (poster_public_key, timestamp DESC);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX poster_public_key_nft_timestamp_idx ON post_entry (poster_public_key, timestamp, is_nft DESC);
		`)
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
 			CREATE INDEX nft_timestamp_idx ON post_entry (timestamp, is_nft DESC);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP INDEX IF EXISTS badger_key_idx;
			DROP INDEX IF EXISTS timestamp_idx;
			DROP INDEX IF EXISTS nft_idx;
			DROP INDEX IF EXISTS reposted_post_hash_idx;
			DROP INDEX IF EXISTS parent_post_hash_idx;
			DROP INDEX IF EXISTS poster_public_key_timestamp_idx;
			DROP INDEX IF EXISTS poster_public_key_nft_timestamp_idx;
			DROP INDEX IF EXISTS nft_timestamp_idx;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
