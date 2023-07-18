package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createProfileEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				public_key                   VARCHAR PRIMARY KEY NOT NULL,
			    pkid						 VARCHAR NOT NULL,
				username                     VARCHAR,
				description                  VARCHAR,
				profile_pic                  BYTEA,
				creator_basis_points         BIGINT NOT NULL,
				coin_watermark_nanos         BIGINT NOT NULL,
				minting_disabled             BOOLEAN NOT NULL,
				dao_coin_minting_disabled    BOOLEAN NOT NULL,
				dao_coin_transfer_restriction_status SMALLINT NOT NULL,
				extra_data                   JSONB,
				badger_key                   BYTEA NOT NULL
			);
			CREATE INDEX {tableName}_pkid_idx ON {tableName} (pkid);
			CREATE INDEX {tableName}_username_idx ON {tableName} (username);
			CREATE INDEX {tableName}_badger_key_idx ON {tableName} (badger_key);
			-- TODO: Define FK relation to post_entry table.
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createProfileEntryTable(db, "profile_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE profile_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
