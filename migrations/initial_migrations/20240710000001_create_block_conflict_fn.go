package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS block_height_unique_idx ON block (height);

			CREATE OR REPLACE FUNCTION handle_block_conflict(block_hash_old varchar, block_hash_new varchar) RETURNS VOID AS $$
			BEGIN
				IF block_hash_old = block_hash_new THEN
        			RETURN;
    			END IF;
				DELETE FROM transaction_partitioned WHERE block_hash = block_hash_old;
				DELETE FROM utxo_operation WHERE block_hash = block_hash_old;
				DELETE FROM block_signer WHERE block_hash = block_hash_old;
				DELETE FROM stake_reward WHERE block_hash = block_hash_old;
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
