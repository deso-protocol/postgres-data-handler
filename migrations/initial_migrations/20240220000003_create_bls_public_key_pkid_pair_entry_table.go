package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createBLSPublicKeyPKIDPairEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				pkid VARCHAR NOT NULL,
				bls_public_key VARCHAR NOT NULL,

				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX {tableName}_pkid_idx ON {tableName} (pkid);
			CREATE INDEX {tableName}_bls_public_key_idx ON {tableName} (bls_public_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createBLSPublicKeyPKIDPairEntryTable(db, "bls_public_key_pkid_pair_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS bls_public_key_pkid_pair_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
