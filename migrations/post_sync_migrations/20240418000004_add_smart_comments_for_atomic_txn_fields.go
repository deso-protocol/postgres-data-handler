package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			comment on view transaction is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName transactions|@fieldName block\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactions|@fieldName account\n@unique transaction_hash\n@foreignKey (wrapper_transaction_hash) references transaction (transaction_hash)|@foreignFieldName innerTransactions|@fieldName wrapperTransaction';
			CREATE INDEX transaction_wrapper_transaction_hash_idx ON transaction_partitioned (wrapper_transaction_hash desc);
			CREATE INDEX transaction_wrapper_transaction_hash_and_idx_in_wrapper_idx ON transaction_partitioned (wrapper_transaction_hash desc, index_in_wrapper_transaction desc);
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP INDEX transaction_wrapper_transaction_hash_idx;
			DROP INDEX transaction_wrapper_transaction_hash_and_idx_in_wrapper_idx;
			comment on view transaction is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName transactions|@fieldName block\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactions|@fieldName account\n@unique transaction_hash';
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
