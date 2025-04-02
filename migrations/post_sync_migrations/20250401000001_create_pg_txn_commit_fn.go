package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create a function that will be called when a transaction is committed.
		// For now, this is empty, but will still be called when a transaction is committed.
		// Other implementations of the postgres-data-handler may overwrite this function.
		// If this function has already been created, we don't want this placeholder to overwrite it.
		_, err := db.Exec(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc 
        WHERE proname = 'on_pdh_pg_txn_committed' 
        AND pg_function_is_visible(oid)
    ) THEN
        CREATE FUNCTION on_pdh_pg_txn_committed()
        RETURNS void AS $BODY$
        BEGIN
            -- Empty function body
        END;
        $BODY$ LANGUAGE plpgsql;
    END IF;
END
$$;
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
