package initial_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		CREATE TABLE transaction_type (
				type  INTEGER PRIMARY KEY,
				name  TEXT NOT NULL
		);

		INSERT INTO transaction_type (type, name) VALUES
				(0, 'Unset'),
				(1, 'Block Reward'),
				(2, 'Basic Transfer'),
				(3, 'Bitcoin Exchange'),
				(4, 'Private Message'),
				(5, 'Submit Post'),
				(6, 'Update Profile'),
				(8, 'Update Bitcoin USD Exchange Rate'),
				(9, 'Follow'),
				(10, 'Like'),
				(11, 'Creator Coin'),
				(12, 'Swap Identity'),
				(13, 'Update Global Params'),
				(14, 'Creator Coin Transfer'),
				(15, 'Create NFT'),
				(16, 'Update NFT'),
				(17, 'Accept NFT Bid'),
				(18, 'NFT Bid'),
				(19, 'NFT Transfer'),
				(20, 'Accept NFT Transfer'),
				(21, 'Burn NFT'),
				(22, 'Authorize Derived Key'),
				(23, 'Messaging Group'),
				(24, 'DAO Coin'),
				(25, 'DAO Coin Transfer'),
				(26, 'DAO Coin Limit Order'),
				(27, 'Create User Association'),
				(28, 'Delete User Association'),
				(29, 'Create Post Association'),
				(30, 'Delete Post Association'),
				(31, 'Access Group'),
				(32, 'Access Group Members'),
				(33, 'New Message');
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS transaction_type;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
