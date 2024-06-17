package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createBlockSignerTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				block_hash   VARCHAR NOT NULL,
				signer_index BIGINT NOT NULL,
				PRIMARY KEY(block_hash, signer_index)
			);
			CREATE INDEX {tableName}_block_hash_idx ON {tableName} (block_hash);
			CREATE INDEX {tableName}_block_hash_signer_index_idx ON {tableName} (block_hash, signer_index);
			CREATE INDEX {tableName}_signer_index_idx ON {tableName} (signer_index);
			create index block_proposer_voting_public_key on block (proposer_voting_public_key);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createBlockSignerTable(db, "block_signer")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS block_signer;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
