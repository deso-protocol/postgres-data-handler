package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := RunMigrationWithRetries(db, `
			CREATE OR REPLACE FUNCTION get_transaction_count(transaction_type integer)
			RETURNS bigint AS
			$BODY$
			DECLARE
				count_value bigint;
				padded_transaction_type varchar;
			BEGIN
				IF transaction_type < 1 OR transaction_type > 44 THEN
					RAISE EXCEPTION '% is not a valid transaction type', transaction_type;
				END IF;
			
				padded_transaction_type := LPAD(transaction_type::text, 2, '0');
			
				EXECUTE format('SELECT COALESCE(NULLIF(COALESCE(reltuples::bigint, 0), -1), 0) FROM pg_class WHERE relname = ''transaction_partition_%s''', padded_transaction_type) INTO count_value;
				RETURN count_value;
			END;
			$BODY$
			LANGUAGE plpgsql
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		err := RunMigrationWithRetries(db, `
			CREATE OR REPLACE FUNCTION get_transaction_count(transaction_type integer)
			RETURNS bigint AS
			$BODY$
			DECLARE
				count_value bigint;
				padded_transaction_type varchar;
			BEGIN
				IF transaction_type < 1 OR transaction_type > 33 THEN
					RAISE EXCEPTION '% is not a valid transaction type', transaction_type;
				END IF;
			
				padded_transaction_type := LPAD(transaction_type::text, 2, '0');
			
				EXECUTE format('SELECT COALESCE(reltuples::bigint, 0) FROM pg_class WHERE relname = ''transaction_partition_%s''', padded_transaction_type) INTO count_value;
				RETURN count_value;
			END;
			$BODY$
			LANGUAGE plpgsql
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
