package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		INSERT INTO transaction_type (type, name) VALUES
				(34, 'Register As Validator'),
				(35, 'Unregister As Validator'),
				(36, 'Stake'),
				(37, 'Unstake'),
				(38, 'Unlock Stake'),
				(39, 'Unjail Validator'),
				(40, 'Coin Lockup'),
				(41, 'Update Coin Lockup Params'),
				(42, 'Coin Lockup Transfer'),
				(43, 'Coin Unlock');
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			delete from transaction_type where type >= 34 AND type <= 43;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
