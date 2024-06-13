package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE INDEX post_association_extra_sender_idx ON post_association_entry ((extra_data ->> 'SenderPublicKey'));
			CREATE INDEX post_association_type_extra_sender_idx ON post_association_entry (association_type, (extra_data ->> 'SenderPublicKey'));
			CREATE INDEX post_association_post_type_extra_sender_idx ON post_association_entry (post_hash, association_type, (extra_data ->> 'SenderPublicKey'));
			CREATE INDEX post_association_extra_receiver_idx ON post_association_entry ((extra_data ->> 'ReceiverPublicKey'));
			CREATE INDEX post_association_type_extra_receiver_idx ON post_association_entry (association_type, (extra_data ->> 'ReceiverPublicKey'));
			
			comment on column post_entry_view.post_hash is E'@notNull';
			comment on column post_association_entry_view.association_type is E'@notNull';
			comment on column post_association_entry_view.association_value is E'@notNull';
			comment on column user_association_entry_view.association_type is E'@notNull';
			comment on column user_association_entry_view.association_value is E'@notNull';
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP INDEX IF EXISTS post_association_extra_sender_idx;
			DROP INDEX IF EXISTS post_association_type_extra_sender_idx;
			DROP INDEX IF EXISTS post_association_post_type_extra_sender_idx;
			DROP INDEX IF EXISTS post_association_extra_receiver_idx;
			DROP INDEX IF EXISTS post_association_type_extra_receiver_idx;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
