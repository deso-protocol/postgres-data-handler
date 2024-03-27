package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
				comment on table jailed_history_event is E'@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName jailedHistoryEvents|@fieldName validatorEntry\n@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName jailedHistoryEvents|@fieldName account';
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
				comment on table jailed_history_event is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
