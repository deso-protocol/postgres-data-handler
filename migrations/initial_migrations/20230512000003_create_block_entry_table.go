package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE block (
				block_hash            			VARCHAR PRIMARY KEY,
				prev_block_hash            		VARCHAR,
				txn_merkle_root            		VARCHAR NOT NULL,
				timestamp		  				timestamp NOT NULL,
				height                			BIGINT NOT NULL,
				nonce                			BIGINT,
				extra_nonce			  			BIGINT,
				badger_key      BYTEA NOT NULL
			);
			CREATE INDEX block_prev_block_hash_idx ON block (prev_block_hash);
			CREATE INDEX block_height_idx ON block (height desc);
			CREATE INDEX block_timestamp_idx ON block (timestamp desc);
			CREATE UNIQUE INDEX block_badger_key_idx ON block (badger_key);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE block;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
