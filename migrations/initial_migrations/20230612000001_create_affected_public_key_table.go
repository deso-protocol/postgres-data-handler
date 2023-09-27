package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE affected_public_key (
				public_key VARCHAR,
				transaction_hash VARCHAR,
				txn_type smallint not null,
				metadata VARCHAR,
				timestamp timestamp,
				PRIMARY KEY(public_key, transaction_hash, metadata)
			);
			CREATE INDEX affected_public_key_public_key_idx ON affected_public_key (public_key);
			CREATE INDEX affected_public_key_metadata_key_idx ON affected_public_key (metadata);
			CREATE INDEX affected_public_key_txn_type_idx ON affected_public_key (txn_type);
			CREATE INDEX affected_public_key_timestamp_idx ON affected_public_key (timestamp desc);
			CREATE INDEX affected_public_key_transaction_hash_idx ON affected_public_key (transaction_hash);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS affected_public_key;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
