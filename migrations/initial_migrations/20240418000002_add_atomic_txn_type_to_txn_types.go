package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		INSERT INTO transaction_type (type, name) VALUES
				(44, 'Atomic Transaction');
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			delete from transaction_type where type = 44;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
