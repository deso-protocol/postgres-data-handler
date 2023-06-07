package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
	"strings"
)

// Create a postgres trigger function to fire on insert in order to update the public_key table.
func createPublicKeyTriggerFn(db *bun.DB, triggerFnName string, fieldName string) error {
	// Create a trigger to run on insert into the public_key table.
	err := RunMigrationWithRetries(db, strings.Replace(strings.Replace(`
 			CREATE OR REPLACE FUNCTION {trigger_fn_name}() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO public_key(public_key) 
				VALUES (NEW.{field_name})
				ON CONFLICT (public_key) DO NOTHING;
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;
		`, "{field_name}", fieldName, -1), "{trigger_fn_name}", triggerFnName, -1))
	return err
}

// Create a postgres trigger to fire on insert into the specified table.
func assignPublicKeyTriggerFn(db *bun.DB, tableName string, triggerName string, triggerFnName string) error {
	err := RunMigrationWithRetries(db, strings.Replace(strings.Replace(strings.Replace(`
 			CREATE TRIGGER {trigger_name}
			AFTER INSERT ON {table_name}
			FOR EACH ROW
			EXECUTE PROCEDURE {trigger_fn_name}();
		`, "{table_name}", tableName, -1), "{trigger_name}", triggerName, -1), "{trigger_fn_name}", triggerFnName, -1))
	return err
}

func createAndAssignPublicKeyTriggerFn(db *bun.DB, tableName string, fieldName string) error {
	triggerFnName := "insert_public_key_" + tableName + "_" + fieldName
	triggerName := "insert_public_key_trigger_" + tableName + "_" + fieldName
	err := createPublicKeyTriggerFn(db, triggerFnName, fieldName)
	if err != nil {
		return err
	}
	err = assignPublicKeyTriggerFn(db, tableName, triggerName, triggerFnName)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create public key table.
		err := RunMigrationWithRetries(db, `
 			CREATE TABLE public_key (
 				public_key VARCHAR PRIMARY KEY
			);
		`)
		if err != nil {
			return err
		}
		// Populate public key table with initial data post-hypersync.
		err = RunMigrationWithRetries(db, `
 			INSERT INTO public_key(public_key)
			select distinct public_key from (
				select public_key from profile_entry
				union all
				select poster_public_key from post_entry
				union all
				select pkid as public_key from deso_balance_entry
				union all
				select hodler_pkid as public_key from balance_entry
				union all
				select creator_pkid as public_key from balance_entry
				union all
				select owner_public_key  as public_key from derived_key_entry
				union all
				select public_key from pkid_entry
				union all
				select sender_public_key as public_key from message_entry
				union all
				select recipient_public_key as public_key from message_entry
				union all
				select sender_access_group_owner_public_key as public_key from new_message_entry
				union all
				select recipient_access_group_owner_public_key as public_key from new_message_entry
				union all
				select target_user_pkid as public_key from user_association_entry
				union all
				select access_group_owner_public_key as public_key from access_group_entry
				union all
				select access_group_member_public_key as public_key from access_group_member_entry
				union all
				select follower_pkid as public_key from follow_entry
				union all
				select public_key from like_entry
			) public_keys;
		`)
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "profile_entry", "public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "pkid_entry", "public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "like_entry", "public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "post_entry", "poster_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "deso_balance_entry", "pkid")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "balance_entry", "hodler_pkid")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "balance_entry", "creator_pkid")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "derived_key_entry", "owner_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "message_entry", "sender_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "message_entry", "recipient_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "new_message_entry", "sender_access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "new_message_entry", "recipient_access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "user_association_entry", "target_user_pkid")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "access_group_entry", "access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "access_group_member_entry", "access_group_owner_public_key")
		if err != nil {
			return err
		}

		err = createAndAssignPublicKeyTriggerFn(db, "follow_entry", "follower_pkid")
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS public_key CASCADE;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
