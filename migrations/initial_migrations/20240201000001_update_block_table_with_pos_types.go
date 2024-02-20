package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func updateBlockTableWithPoSFields(db *bun.DB, tableName string) error {
	_, err := db.Exec(`
			ALTER TABLE block
			ADD COLUMN block_version BIGINT,
			ADD COLUMN txn_connect_status_by_index_hash VARCHAR,
			ADD COLUMN proposer_voting_public_key VARCHAR,
			ADD COLUMN proposer_random_seed_signature VARCHAR,
			ADD COLUMN proposed_in_view BIGINT,
			ADD COLUMN proposer_vote_partial_signature VARCHAR;
		`)
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return updateBlockTableWithPoSFields(db, "block")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE block
			DROP COLUMN block_version,
			DROP COLUMN txn_connect_status_by_index_hash,
			DROP COLUMN proposer_voting_public_key,
			DROP COLUMN proposer_random_seed_signature,
			DROP COLUMN proposed_in_view,
			DROP COLUMN proposer_vote_partial_signature;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
