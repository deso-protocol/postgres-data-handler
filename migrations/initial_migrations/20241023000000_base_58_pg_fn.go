package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE FUNCTION base58_check_encode_with_prefix(input bytea) RETURNS TEXT AS $$
			DECLARE
				prefix bytea := E'\\xcd1400'::bytea;
				b bytea;
				big_val NUMERIC;
			BEGIN
				b := prefix || input || checksum(prefix || input);
			
				-- Convert bytea to a big numeric for Base58 encoding
				SELECT INTO big_val bytes_to_bigint(b);
				RETURN base58_encode(big_val);
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;
		`)
		if err != nil {
			return err
		}
		return nil

	}, func(ctx context.Context, db *bun.DB) error {
		// Lastly, delete the noaccess role.
		_, err := db.Exec(`
			DROP FUNCTION base58_check_encode_with_prefix(bytea);
`)
		if err != nil {
			return err
		}

		return nil
	})
}
