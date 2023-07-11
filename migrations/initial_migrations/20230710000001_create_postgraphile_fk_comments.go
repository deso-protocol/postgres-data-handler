package initial_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			comment on table access_group_entry is E'@foreignKey (access_group_owner_public_key) references public_key (public_key)';
			comment on table access_group_member_entry is E'@foreignKey (access_group_member_public_key) references public_key (public_key)';
			comment on table affected_public_key is E'@foreignKey (public_key) references public_key (public_key)\n@foreignKey (transaction_hash) references transaction (transaction_hash)';
			comment on table balance_entry is E'@foreignKey (hodler_pkid) references pkid (pkid)\n@foreignKey (creator_pkid) references pkid (pkid)';
			comment on table derived_key_entry is E'@foreignKey (owner_public_key) references public_key (public_key)';
			comment on table deso_balance_entry is E'@foreignKey (pkid) references pkid (pkid)';
			comment on table diamond_entry is E'@foreignKey (sender_pkid) references pkid (pkid)\n@foreignKey (receiver_pkid) references pkid (pkid)\n@foreignKey (post_hash) references post_entry (post_hash)';
			comment on table follow_entry is E'@foreignKey (follower_pkid) references pkid (pkid)\n@foreignKey (followed_pkid) references pkid (pkid)';
			comment on table like_entry is E'@foreignKey (public_key) references public_key (public_key)\n@foreignKey (post_hash) references post_entry (post_hash)';
			comment on table message_entry is E'@foreignKey (sender_public_key) references public_key (public_key)\n@foreignKey (recipient_public_key) references public_key (public_key)';
			comment on table new_message_entry is E'@foreignKey (sender_access_group_owner_public_key) references public_key (public_key)\n@foreignKey (recipient_access_group_owner_public_key) references public_key (public_key)\n@foreignKey (sender_access_group_public_key) references access_group_entry (access_group_public_key)\n@foreignKey (recipient_access_group_public_key) references access_group_entry (access_group_public_key)';
			comment on table nft_bid_entry is E'@foreignKey (bidder_pkid) references pkid (pkid)\n@foreignKey (nft_post_hash) references post_entry (post_hash)\n@foreignKey (accepted_block_height) references block (height)';
			comment on table nft_entry is E'@foreignKey (last_owner_pkid) references pkid (pkid)\n@foreignKey (owner_pkid) references pkid (pkid)\n@foreignKey (nft_post_hash) references post_entry (post_hash)';
			comment on table post_association_entry is E'@foreignKey (transactor_pkid) references pkid (pkid)\n@foreignKey (app_pkid) references pkid (pkid)\n@foreignKey (post_hash) references post_entry (post_hash)\n@foreignKey (block_height) references block (height)';
			comment on table post_association_entry is E'@foreignKey (transactor_pkid) references pkid (pkid)\n@foreignKey (app_pkid) references pkid (pkid)\n@foreignKey (post_hash) references post_entry (post_hash)\n@foreignKey (block_height) references block (height)';
			comment on table post_entry is E'@foreignKey (poster_public_key) references public_key (public_key)\n@foreignKey (parent_post_hash) references post_entry (post_hash)\n@foreignKey (reposted_post_hash) references post_entry (post_hash)';
			comment on table profile_entry is E'@foreignKey (pkid) references pkid (pkid)\n@foreignKey (public_key) references public_key (public_key)';
			comment on table transaction is E'@foreignKey (block_hash) references block (block_hash)';
			comment on table user_association_entry is E'@foreignKey (transactor_pkid) references pkid (pkid)\n@foreignKey (app_pkid) references pkid (pkid)\n@foreignKey (target_user_pkid) references pkid (pkid)\n@foreignKey (block_height) references block (height)';
			comment on table utxo_operation is E'@foreignKey (block_hash, transaction_index) references transaction (block_hash, index_in_block)';
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			comment on table access_group_entry is NULL;
			comment on table access_group_member_entry is NULL;
			comment on table affected_public_key is NULL;
			comment on table balance_entry is NULL;
			comment on table derived_key_entry is NULL;
			comment on table deso_balance_entry is NULL;
			comment on table diamond_entry is NULL;
			comment on table follow_entry is NULL;
			comment on table like_entry is NULL;
			comment on table message_entry is NULL;
			comment on table new_message_entry is NULL;
			comment on table nft_bid_entry is NULL;
			comment on table nft_entry is NULL;
			comment on table post_association_entry is NULL;
			comment on table post_association_entry is NULL;
			comment on table post_entry is NULL;
			comment on table profile_entry is NULL;
			comment on table transaction is NULL;
			comment on table user_association_entry is NULL;
			comment on table utxo_operation is NULL;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
