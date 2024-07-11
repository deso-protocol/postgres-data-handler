package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS block_height_unique_idx ON block (height);

			CREATE OR REPLACE FUNCTION handle_block_conflict(block_hash_val varchar) RETURNS VOID AS $$
			BEGIN
				DELETE FROM transaction_partitioned WHERE block_hash = block_hash_val;
				DELETE FROM utxo_operation WHERE block_hash = block_hash_val;
				DELETE FROM block_signer WHERE block_hash = block_hash_val;
			END;
			$$ LANGUAGE plpgsql;
		`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP INDEX IF EXISTS block_height_unique_idx;
			DROP function IF EXISTS handle_block_conflict;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
