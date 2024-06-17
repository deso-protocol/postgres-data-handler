package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			create or replace view wallet as
			select case when pkid_entry.pkid is null then public_key.public_key else pkid_entry.pkid end as pkid, public_key.public_key as public_key
			from public_key
			left join pkid_entry
			on pkid_entry.public_key = public_key.public_key;
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE OR REPLACE VIEW wallet AS
			SELECT pkid, public_key FROM pkid_entry
			UNION ALL
			SELECT public_key AS pkid, public_key
			FROM public_key
			WHERE public_key NOT IN (SELECT public_key FROM pkid_entry)
			AND public_key NOT IN (SELECT pkid FROM pkid_entry);
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
