package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE transaction_partitioned ALTER COLUMN index_in_block DROP NOT NULL;
			ALTER TABLE transaction_partitioned ADD COLUMN wrapper_transaction_hash VARCHAR;
			ALTER TABLE transaction_partitioned ADD COLUMN index_in_wrapper_transaction BIGINT;
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE OR REPLACE VIEW transaction AS
			SELECT * FROM transaction_partitioned;
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DELETE FROM transaction_partitioned where index_in_block IS NULL;
			ALTER TABLE transaction_partitioned ALTER COLUMN index_in_block SET NOT NULL;
			ALTER TABLE transaction_partitioned DROP COLUMN wrapper_transaction_hash CASCADE;
			ALTER TABLE transaction_partitioned DROP COLUMN index_in_wrapper_transaction CASCADE;
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE OR REPLACE VIEW transaction AS
			SELECT * FROM transaction_partitioned;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
