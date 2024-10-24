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
				deso_locked_nanos			 BIGINT NOT NULL,
				cc_coins_in_circulation_nanos BIGINT NOT NULL,
				dao_coins_in_circulation_nanos_hex VARCHAR NOT NULL,
				dao_coin_minting_disabled    BOOLEAN NOT NULL,
				dao_coin_transfer_restriction_status SMALLINT NOT NULL,
				extra_data                   JSONB,
				badger_key                   BYTEA NOT NULL,
				coin_price_deso_nanos NUMERIC
				GENERATED ALWAYS AS (
						CASE
						WHEN cc_coins_in_circulation_nanos = 0 THEN 0
						ELSE
								(ROUND((
						deso_locked_nanos::NUMERIC / (cc_coins_in_circulation_nanos::NUMERIC * 0.33333) * 1e9
				)::NUMERIC, 0))::NUMERIC
						END
				) STORED
			);
			CREATE INDEX {tableName}_pkid_idx ON {tableName} (pkid);
			CREATE INDEX {tableName}_username_idx ON {tableName} (username);
			CREATE INDEX {tableName}_coin_price_idx ON {tableName} (coin_price_deso_nanos desc);
			CREATE INDEX {tableName}_username_lower_idx ON {tableName} (LOWER(username));
			CREATE INDEX {tableName}_username_ilike_idx ON {tableName} (LOWER("username"));
			CREATE INDEX {tableName}_username_gin_idx ON {tableName} USING gin("username" gin_trgm_ops);


			CREATE INDEX {tableName}_badger_key_idx ON {tableName} (badger_key);
		`, "{tableName}", tableName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createProfileEntryTable(db, "profile_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS profile_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
