package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

func createDaoCoinLimitOrderEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
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
				badger_key BYTEA PRIMARY KEY
			);
		
		CREATE INDEX {tableName}_order_id_idx ON {tableName} (order_id);
		CREATE INDEX {tableName}_transactor_pkid_idx ON {tableName} (transactor_pkid);
		CREATE INDEX {tableName}_buying_pkid_idx ON {tableName} (buying_dao_coin_creator_pkid);
		CREATE INDEX {tableName}_selling_pkid_idx ON {tableName} (selling_dao_coin_creator_pkid);
		CREATE INDEX {tableName}_operation_type_idx ON {tableName} (operation_type);
		CREATE INDEX {tableName}_fill_type_idx ON {tableName} (fill_type);
		CREATE INDEX {tableName}_block_height_idx ON {tableName} (block_height desc);
		`, "{tableName}", tableName, -1), nil)
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createPostAssociationEntryTable(db, "dao_coin_limit_order_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE dao_coin_limit_order_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
