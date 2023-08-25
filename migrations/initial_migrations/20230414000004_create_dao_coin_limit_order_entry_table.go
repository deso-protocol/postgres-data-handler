package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createDaoCoinLimitOrderEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`

			CREATE OR REPLACE FUNCTION hex_to_numeric(hex_string TEXT) RETURNS NUMERIC AS $$
			DECLARE
				result NUMERIC := 0;
				current_char CHAR(1);
				current_value INT;
			BEGIN
				FOR i IN 1 .. LENGTH(hex_string) LOOP
					current_char := SUBSTRING(hex_string, i, 1);
			
					IF '0' <= current_char AND current_char <= '9' THEN
						current_value := ASCII(current_char) - ASCII('0');
					ELSIF 'A' <= upper(current_char) AND upper(current_char) <= 'F' THEN
						current_value := ASCII(upper(current_char)) - ASCII('A') + 10;
					END IF;
			
					result := result * 16 + current_value;
				END LOOP;
			
				RETURN result;
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;

			CREATE TABLE {tableName} (
				order_id VARCHAR,
				transactor_pkid VARCHAR,
				buying_dao_coin_creator_pkid VARCHAR,
				selling_dao_coin_creator_pkid VARCHAR,
				scaled_exchange_rate_coins_to_sell_per_coin_to_buy_hex VARCHAR,
				quantity_to_fill_in_base_units_hex VARCHAR,
				operation_type INTEGER,
				fill_type INTEGER,
				block_height BIGINT,
				badger_key BYTEA PRIMARY KEY,
				is_dao_coin_const BOOLEAN DEFAULT TRUE NOT NULL,
				scaled_exchange_rate_coins_to_sell_per_coin_to_buy_numeric NUMERIC
				GENERATED ALWAYS AS (
					hex_to_numeric(substring(scaled_exchange_rate_coins_to_sell_per_coin_to_buy_hex from 3))
				) STORED,
				quantity_to_fill_in_base_units_numeric NUMERIC
				GENERATED ALWAYS AS (
					hex_to_numeric(substring(quantity_to_fill_in_base_units_hex from 3))
				) STORED
			);
		
		CREATE INDEX {tableName}_order_id_idx ON {tableName} (order_id);
		CREATE INDEX {tableName}_transactor_pkid_idx ON {tableName} (transactor_pkid);
		CREATE INDEX {tableName}_buying_pkid_idx ON {tableName} (buying_dao_coin_creator_pkid);
		CREATE INDEX {tableName}_selling_pkid_idx ON {tableName} (selling_dao_coin_creator_pkid);
		CREATE INDEX {tableName}_operation_type_idx ON {tableName} (operation_type);
		CREATE INDEX {tableName}_is_dao_coin_const_idx ON {tableName} (is_dao_coin_const);
		CREATE INDEX {tableName}_fill_type_idx ON {tableName} (fill_type);
		CREATE INDEX {tableName}_block_height_idx ON {tableName} (block_height desc);
		CREATE INDEX {tableName}_scaled_exchange_rate_idx ON {tableName} (scaled_exchange_rate_coins_to_sell_per_coin_to_buy_numeric desc);
		CREATE INDEX {tableName}_quantity_to_fill_idx ON {tableName} (quantity_to_fill_in_base_units_numeric desc);
		`, "{tableName}", tableName, -1), nil)
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createDaoCoinLimitOrderEntryTable(db, "dao_coin_limit_order_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS dao_coin_limit_order_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
