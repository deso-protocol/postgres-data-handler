package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE transaction (
				transaction_hash       			VARCHAR PRIMARY KEY,
				block_hash            			VARCHAR NOT NULL,
				version            				SMALLINT NOT NULL,
				inputs                			JSONB,
				outputs                			JSONB,
				fee_nanos              			BIGINT,
				nonce_experiation_block_height  BIGINT,
				nonce_partial_id                BIGINT,
				txn_meta                		JSONB,
				txn_meta_bytes			  		BYTEA,
				tx_index_metadata				JSONB,
				tx_index_basic_transfer_metadata JSONB,
				txn_type						SMALLINT NOT NULL,
				public_key                		VARCHAR,
				extra_data                		JSONB,
				signature		              	BYTEA,
				txn_bytes						BYTEA NOT NULL,
				index_in_block		  			INTEGER NOT NULL
			);
			CREATE INDEX transaction_index_in_block_idx ON transaction (index_in_block);
			CREATE INDEX transaction_block_hash_index_idx ON transaction (block_hash, index_in_block);
			CREATE INDEX transaction_type_idx ON transaction (txn_type);
			CREATE INDEX transaction_public_key_idx ON transaction (public_key);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE transaction;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
