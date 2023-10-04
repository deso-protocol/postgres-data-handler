package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE transaction_partitioned
			(
				transaction_hash                 varchar  not null,
				transaction_id	                 varchar  not null,
				block_hash                       varchar  not null,
				version                          smallint not null,
				inputs                           jsonb,
				outputs                          jsonb,
				fee_nanos                        bigint,
				nonce_expiration_block_height    bigint,
				nonce_partial_id                 bigint,
				txn_meta                         jsonb,
				txn_meta_bytes                   bytea,
				tx_index_metadata                jsonb,
				tx_index_basic_transfer_metadata jsonb,
				txn_type                         smallint not null,
				public_key                       varchar,
				extra_data                       jsonb,
				signature                        bytea,
				txn_bytes                        bytea    not null,
				index_in_block                   integer  not null,
				block_height    				 bigint,
				timestamp                        timestamp,
				badger_key                       bytea    not null,
				PRIMARY KEY (transaction_hash, txn_type)
			) PARTITION BY LIST (txn_type);
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE TABLE transaction_partition_01 PARTITION OF transaction_partitioned FOR VALUES IN (1);
			CREATE TABLE transaction_partition_02 PARTITION OF transaction_partitioned FOR VALUES IN (2);
			CREATE TABLE transaction_partition_03 PARTITION OF transaction_partitioned FOR VALUES IN (3);
			CREATE TABLE transaction_partition_04 PARTITION OF transaction_partitioned FOR VALUES IN (4);
			CREATE TABLE transaction_partition_05 PARTITION OF transaction_partitioned FOR VALUES IN (5);
			CREATE TABLE transaction_partition_06 PARTITION OF transaction_partitioned FOR VALUES IN (6);
			CREATE TABLE transaction_partition_07 PARTITION OF transaction_partitioned FOR VALUES IN (7);
			CREATE TABLE transaction_partition_08 PARTITION OF transaction_partitioned FOR VALUES IN (8);
			CREATE TABLE transaction_partition_09 PARTITION OF transaction_partitioned FOR VALUES IN (9);
			CREATE TABLE transaction_partition_10 PARTITION OF transaction_partitioned FOR VALUES IN (10);
			CREATE TABLE transaction_partition_11 PARTITION OF transaction_partitioned FOR VALUES IN (11);
			CREATE TABLE transaction_partition_12 PARTITION OF transaction_partitioned FOR VALUES IN (12);
			CREATE TABLE transaction_partition_13 PARTITION OF transaction_partitioned FOR VALUES IN (13);
			CREATE TABLE transaction_partition_14 PARTITION OF transaction_partitioned FOR VALUES IN (14);
			CREATE TABLE transaction_partition_15 PARTITION OF transaction_partitioned FOR VALUES IN (15);
			CREATE TABLE transaction_partition_16 PARTITION OF transaction_partitioned FOR VALUES IN (16);
			CREATE TABLE transaction_partition_17 PARTITION OF transaction_partitioned FOR VALUES IN (17);
			CREATE TABLE transaction_partition_18 PARTITION OF transaction_partitioned FOR VALUES IN (18);
			CREATE TABLE transaction_partition_19 PARTITION OF transaction_partitioned FOR VALUES IN (19);
			CREATE TABLE transaction_partition_20 PARTITION OF transaction_partitioned FOR VALUES IN (20);
			CREATE TABLE transaction_partition_21 PARTITION OF transaction_partitioned FOR VALUES IN (21);
			CREATE TABLE transaction_partition_22 PARTITION OF transaction_partitioned FOR VALUES IN (22);
			CREATE TABLE transaction_partition_23 PARTITION OF transaction_partitioned FOR VALUES IN (23);
			CREATE TABLE transaction_partition_24 PARTITION OF transaction_partitioned FOR VALUES IN (24);
			CREATE TABLE transaction_partition_25 PARTITION OF transaction_partitioned FOR VALUES IN (25);
			CREATE TABLE transaction_partition_26 PARTITION OF transaction_partitioned FOR VALUES IN (26);
			CREATE TABLE transaction_partition_27 PARTITION OF transaction_partitioned FOR VALUES IN (27);
			CREATE TABLE transaction_partition_28 PARTITION OF transaction_partitioned FOR VALUES IN (28);
			CREATE TABLE transaction_partition_29 PARTITION OF transaction_partitioned FOR VALUES IN (29);
			CREATE TABLE transaction_partition_30 PARTITION OF transaction_partitioned FOR VALUES IN (30);
			CREATE TABLE transaction_partition_31 PARTITION OF transaction_partitioned FOR VALUES IN (31);
			CREATE TABLE transaction_partition_32 PARTITION OF transaction_partitioned FOR VALUES IN (32);
			CREATE TABLE transaction_partition_33 PARTITION OF transaction_partitioned FOR VALUES IN (33);
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE VIEW transaction AS
			SELECT * FROM transaction_partitioned;
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS transaction;
			DROP TABLE IF EXISTS transaction_partitioned;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
