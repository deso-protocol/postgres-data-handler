package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE user_association_entry
				ADD COLUMN payment_uuid VARCHAR GENERATED ALWAYS AS (extra_data ->> 'PaymentUUID') STORED;

			CREATE INDEX user_association_entry_payment_uuid_idx ON user_association_entry (payment_uuid);

			ALTER TABLE transaction_partitioned
				ADD COLUMN payment_uuid VARCHAR GENERATED ALWAYS AS (extra_data ->> 'PaymentUUID') STORED;

			CREATE OR REPLACE VIEW transaction as select * from transaction_partitioned;
			
			CREATE OR REPLACE VIEW payment AS
				SELECT
					association_id,
					access_group_owner_public_key,
					access_group_key_name,
					app_pkid,
					subscription_payment_timestamp as created_at,
					payment_sender_public_key as sender_public_key,
					payment_recipient_public_key as recipient_public_key,
					extra_data,
					payment_uuid
				FROM
					user_association_entry
				WHERE
					association_type = 'PAYMENT' AND
					(extra_data->>'PaymentTstampNanos') IS NOT NULL;
			
			COMMENT ON VIEW payment IS E'@primaryKey association_id\n@unique association_id\n@foreignKey (sender_public_key) references account (public_key)|@foreignFieldName payments|@fieldName sender\n@foreignKey (recipient_public_key) references account (public_key)|@foreignFieldName paymentsReceived|@fieldName recipient\n@foreignKey (access_group_owner_public_key, access_group_key_name) references access_group_entry (access_group_owner_public_key, access_group_key_name)|@foreignFieldName payments|@fieldName accessGroup\n@foreignKey (access_group_owner_public_key, access_group_key_name, sender_public_key) references subscription (access_group_owner_public_key, access_group_key_name, subscriber_public_key)|@foreignFieldName payments|@fieldName subscription\n@foreignKey (payment_uuid) references transaction (payment_uuid)|@foreignFieldName payments|@fieldName paymentTransaction';
			COMMENT ON VIEW user_association_entry_view is E'@name user_association\n@unique association_id\n@unique transactor_pkid, target_user_pkid, app_pkid, association_type, association_value\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTransactor|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName userAssociationsAsAppOwner|@fieldName app\n@foreignKey (target_user_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTarget|@fieldName target\n@foreignKey (block_height) references block (height)|@foreignFieldName userAssociations|@fieldName block\n@foreignKey (payment_uuid) references transaction (payment_uuid)|@foreignFieldName paymentUserAssociations|@fieldName paymentTransaction';		
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP VIEW payment;
			DROP INDEX user_association_entry_payment_uuid_idx;
			ALTER TABLE user_association_entry DROP COLUMN payment_uuid;
			ALTER TABLE transaction_partitioned DROP COLUMN payment_uuid;
			CREATE OR REPLACE VIEW transaction as select * from transaction_partitioned;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
