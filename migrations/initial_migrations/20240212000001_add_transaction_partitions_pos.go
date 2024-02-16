package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE transaction_partition_34 PARTITION OF transaction_partitioned FOR VALUES IN (34);
			CREATE TABLE transaction_partition_35 PARTITION OF transaction_partitioned FOR VALUES IN (35);
			CREATE TABLE transaction_partition_36 PARTITION OF transaction_partitioned FOR VALUES IN (36);
			CREATE TABLE transaction_partition_37 PARTITION OF transaction_partitioned FOR VALUES IN (37);
			CREATE TABLE transaction_partition_38 PARTITION OF transaction_partitioned FOR VALUES IN (38);
			CREATE TABLE transaction_partition_39 PARTITION OF transaction_partitioned FOR VALUES IN (39);
			CREATE TABLE transaction_partition_40 PARTITION OF transaction_partitioned FOR VALUES IN (40);
			CREATE TABLE transaction_partition_41 PARTITION OF transaction_partitioned FOR VALUES IN (41);
			CREATE TABLE transaction_partition_42 PARTITION OF transaction_partitioned FOR VALUES IN (42);
			CREATE TABLE transaction_partition_43 PARTITION OF transaction_partitioned FOR VALUES IN (43);	
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS transaction_partition_34;
			DROP TABLE IF EXISTS transaction_partition_35;
			DROP TABLE IF EXISTS transaction_partition_36;
			DROP TABLE IF EXISTS transaction_partition_37;
			DROP TABLE IF EXISTS transaction_partition_38;
			DROP TABLE IF EXISTS transaction_partition_39;
			DROP TABLE IF EXISTS transaction_partition_40;
			DROP TABLE IF EXISTS transaction_partition_41;
			DROP TABLE IF EXISTS transaction_partition_42;
			DROP TABLE IF EXISTS transaction_partition_43;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
