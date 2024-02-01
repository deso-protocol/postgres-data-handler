package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func updateTransactionTableWithPoSFields(db *bun.DB) error {
	_, err := db.Exec(`
			ALTER TABLE transaction_partitioned
			ADD COLUMN connects BOOLEAN DEFAULT TRUE NOT NULL;
		`)
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return updateTransactionTableWithPoSFields(db)
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE transaction_partitioned
			DROP COLUMN connects;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
