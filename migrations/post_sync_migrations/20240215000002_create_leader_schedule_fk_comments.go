package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
                comment on table leader_schedule_entry is E'@foreignKey (validator_pkid) references account (pkid)|@foreignFieldName leaderScheduleEntries|@fieldName leaderAccount\n@foreignKey (validator_pkid) references validator_entry (validator_pkid)|@foreignFieldName leaderScheduleEntries|@fieldName validatorEntry\n@foreignKey (snapshot_at_epoch_number) references epoch_entry (snapshot_at_epoch_number)|@foreignFieldName leaderScheduleEntries|@fieldName epochEntryBySnapshot';
                comment on column leader_schedule_entry.badger_key is E'@omit';
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
				comment on table leader_schedule_entry is NULL;
                comment on column leader_schedule_entry.badger_key is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
