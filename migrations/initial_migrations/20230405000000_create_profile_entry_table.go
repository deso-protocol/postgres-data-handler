package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		_, err := db.Exec(`
			CREATE TABLE profile_entry (
				public_key                   VARCHAR PRIMARY KEY NOT NULL,
			    pkid						 BYTEA NOT NULL,
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
			CREATE INDEX profile_pkid_idx ON profile_entry (pkid);
			CREATE INDEX profile_username_idx ON profile_entry (username);
			CREATE INDEX profile_badger_key_idx ON profile_entry (badger_key);
			-- TODO: Define FK relation to post_entry table.
		`)
		if err != nil {
			return err
		}
		return nil
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
