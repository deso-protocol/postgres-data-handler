package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
	"strings"
)

// Create a postgres trigger function to fire on insert in order to update the public_key table.
func updatePublicKeyTriggerFn(db *bun.DB, tableName string, fieldName string) error {
	triggerFnName := "insert_public_key_" + tableName + "_" + fieldName
	// Create a trigger to run on insert into the public_key table.
	err := RunMigrationWithRetries(db, strings.Replace(strings.Replace(`
 			CREATE OR REPLACE FUNCTION {trigger_fn_name}() RETURNS TRIGGER AS $$
			DECLARE
				row_count INT;
			BEGIN
				INSERT INTO public_key(public_key) 
				VALUES (NEW.{field_name})
				ON CONFLICT (public_key) DO NOTHING;

				
				GET DIAGNOSTICS row_count = ROW_COUNT;
			
				IF row_count > 0 THEN
					-- No conflict occurred, perform another insert into the wallet table
					-- Only insert into wallet if NEW.{field_name} does not match either a public_key or pkid in the wallet table
					IF NOT EXISTS (
						SELECT 1 
						FROM wallet 
						WHERE public_key = NEW.{field_name} OR pkid = NEW.{field_name}
					) THEN
						INSERT INTO wallet(public_key, pkid)
						VALUES (NEW.{field_name}, NEW.{field_name});
					END IF;
				END IF;

				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;
		`, "{field_name}", fieldName, -1), "{trigger_fn_name}", triggerFnName, -1))
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		// Create new wallet table, migrate to it, and drop the old view.
		err := RunMigrationWithRetries(db, `
		CREATE TABLE wallet_table (
			pkid VARCHAR,
			public_key VARCHAR,
			PRIMARY KEY (public_key)
		);

		insert into wallet_table (pkid, public_key) select pkid, public_key from wallet;
		
		CREATE INDEX idx_wallet_public_key ON wallet_table (public_key);
		CREATE INDEX idx_wallet_pkid ON wallet_table (pkid);

		CREATE OR REPLACE VIEW account AS
			SELECT
				wallet.pkid,
				wallet.public_key,
				profile_entry.username,
				profile_entry.description,
				profile_entry.profile_pic,
				profile_entry.creator_basis_points,
				profile_entry.coin_watermark_nanos,
				profile_entry.minting_disabled,
				profile_entry.dao_coin_minting_disabled,
				profile_entry.dao_coin_transfer_restriction_status,
				profile_entry.extra_data,
				profile_entry.coin_price_deso_nanos,
				profile_entry.deso_locked_nanos,
				profile_entry.cc_coins_in_circulation_nanos,
        		profile_entry.dao_coins_in_circulation_nanos_hex,
				true as token_balance_join_field,
                false as cc_balance_join_field
			FROM
				wallet_table wallet
			LEFT JOIN
				profile_entry
			ON
				wallet.public_key = profile_entry.public_key;

		DROP VIEW IF EXISTS wallet;
		ALTER TABLE wallet_table RENAME TO wallet;

		CREATE OR REPLACE VIEW account AS
			SELECT
				wallet.pkid,
				wallet.public_key,
				profile_entry.username,
				profile_entry.description,
				profile_entry.profile_pic,
				profile_entry.creator_basis_points,
				profile_entry.coin_watermark_nanos,
				profile_entry.minting_disabled,
				profile_entry.dao_coin_minting_disabled,
				profile_entry.dao_coin_transfer_restriction_status,
				profile_entry.extra_data,
				profile_entry.coin_price_deso_nanos,
				profile_entry.deso_locked_nanos,
				profile_entry.cc_coins_in_circulation_nanos,
        		profile_entry.dao_coins_in_circulation_nanos_hex,
				true as token_balance_join_field,
                false as cc_balance_join_field
			FROM
				wallet wallet
			LEFT JOIN
				profile_entry
			ON
				wallet.public_key = profile_entry.public_key;

			comment on view account is E'@unique username\n@unique public_key\n@unique pkid\n@primaryKey public_key';
			    
			comment on table wallet is E'@omit';
		`)

		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "profile_entry", "public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "pkid_entry", "public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "like_entry", "public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "post_entry", "poster_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "deso_balance_entry", "public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "balance_entry", "hodler_pkid")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "balance_entry", "creator_pkid")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "derived_key_entry", "owner_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "message_entry", "sender_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "message_entry", "recipient_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "new_message_entry", "sender_access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "new_message_entry", "recipient_access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "user_association_entry", "target_user_pkid")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "access_group_entry", "access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "access_group_member_entry", "access_group_member_public_key")
		if err != nil {
			return err
		}

		err = updatePublicKeyTriggerFn(db, "follow_entry", "follower_pkid")
		if err != nil {
			return err
		}

		err = RunMigrationWithRetries(db, `
			CREATE OR REPLACE FUNCTION insert_into_wallet_table_from_pkid_entry()
			RETURNS TRIGGER AS $$
			DECLARE
				row_count INT;
			BEGIN
				INSERT INTO wallet (pkid, public_key)
				VALUES (NEW.pkid, NEW.public_key)
				ON CONFLICT (public_key)
				DO UPDATE SET pkid = EXCLUDED.pkid
				WHERE wallet_table.pkid <> EXCLUDED.pkid;
			
				GET DIAGNOSTICS row_count = ROW_COUNT;
			
				IF row_count = 0 THEN
					-- If there was a conflict on public_key, make sure no other public_key is associated with this pkid.
					DELETE FROM wallet WHERE pkid = NEW.pkid AND public_key <> NEW.public_key;
				END IF;

				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;

			CREATE TRIGGER insert_into_wallet_table_from_pkid_entry_trigger
			AFTER INSERT ON pkid_entry
			FOR EACH ROW
			EXECUTE FUNCTION insert_into_wallet_table_from_pkid_entry();
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE wallet cascade;

			CREATE OR REPLACE VIEW wallet AS
			SELECT pkid, public_key FROM pkid_entry
			UNION ALL
			SELECT public_key AS pkid, public_key
			FROM public_key
			WHERE public_key NOT IN (SELECT public_key FROM pkid_entry)
			AND public_key NOT IN (SELECT pkid FROM pkid_entry);

			CREATE OR REPLACE VIEW account AS
			SELECT
				wallet.pkid,
				wallet.public_key,
				profile_entry.username,
				profile_entry.description,
				profile_entry.profile_pic,
				profile_entry.creator_basis_points,
				profile_entry.coin_watermark_nanos,
				profile_entry.minting_disabled,
				profile_entry.dao_coin_minting_disabled,
				profile_entry.dao_coin_transfer_restriction_status,
				profile_entry.extra_data,
				profile_entry.coin_price_deso_nanos,
				profile_entry.deso_locked_nanos,
				profile_entry.cc_coins_in_circulation_nanos,
        		profile_entry.dao_coins_in_circulation_nanos_hex,
				true as token_balance_join_field,
                false as cc_balance_join_field
			FROM
				wallet
			LEFT JOIN
				profile_entry
			ON
				wallet.public_key = profile_entry.public_key;

			comment on view account is E'@unique username\n@unique public_key\n@unique pkid\n@primaryKey public_key';
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
