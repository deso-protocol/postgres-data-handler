package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

// TODO: revisit access group relationships when we refactor the messaging app to use the graphql API.
func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			comment on view account is E'@unique username\n@unique public_key\n@unique pkid\n@primaryKey public_key';
			comment on table access_group_entry is E'@name access_group\n@foreignKey (access_group_owner_public_key) references account (public_key)|@foreignFieldName accessGroupsOwned|@fieldName owner';
			comment on table access_group_member_entry is E'@name access_group_member\n@foreignKey (access_group_member_public_key) references account (public_key)|@foreignFieldName accessGroupMemberships|@fieldName member';
			comment on table affected_public_key is E'@foreignKey (public_key) references account (public_key)|@foreignFieldName transactionHashes|@fieldName account\n@foreignKey (transaction_hash) references transaction (transaction_hash)|@fieldName transaction';
			comment on table balance_entry is E'@name tokenBalance\n@foreignKey (hodler_pkid) references account (pkid)|@foreignFieldName tokenBalances|@fieldName holder\n@foreignKey (creator_pkid) references account (pkid)|@foreignFieldName tokenBalancesAsCreator|@fieldName creator';
			comment on table derived_key_entry is E'@name derived_key\n@foreignKey (owner_public_key) references account (public_key)|@foreignFieldName derivedKeys|@fieldName owner';
			comment on table deso_balance_entry is E'@name desoBalance\n@foreignKey (pkid) references account (public_key)|@fieldName desoBalanceEntry';
			comment on table diamond_entry is E'@name diamond\n@foreignKey (sender_pkid) references account (pkid)|@foreignFieldName diamondsSent|@fieldName sender\n@foreignKey (receiver_pkid) references account (pkid)|@foreignFieldName diamondsReceived|@fieldName reciever\n@foreignKey (post_hash) references post_entry (post_hash)|@foreignFieldName diamonds|@fieldName post';
			comment on table follow_entry is E'@name follow\n@foreignKey (follower_pkid) references account (pkid)|@foreignFieldName following|@fieldName follower\n@foreignKey (followed_pkid) references account (pkid)|@foreignFieldName followers|@fieldName followee';
			comment on table like_entry is E'@name like\n@foreignKey (public_key) references account (public_key)|@foreignFieldName likes|@fieldName account\n@foreignKey (post_hash) references post_entry (post_hash)|@foreignFieldName likes|@fieldName post';
			comment on table message_entry is E'@name legacyMessage\n@foreignKey (sender_public_key) references account (public_key)|@foreignFieldName legacyMessagesSent|@fieldName sender\n@foreignKey (recipient_public_key) references account (public_key)|@foreignFieldName legacyMessagesReceived|@fieldName receiver';
			comment on table new_message_entry is E'@name message\n@foreignKey (sender_access_group_owner_public_key) references account (public_key)|@foreignFieldName messagesSent|@fieldName sender\n@foreignKey (recipient_access_group_owner_public_key) references account (public_key)|@foreignFieldName messagesReceived|@fieldName receiver\n@foreignKey (sender_access_group_public_key) references access_group_entry (access_group_public_key)|@foreignFieldName groupMessagesSent|@fieldName senderAccessGroup\n@foreignKey (recipient_access_group_public_key) references access_group_entry (access_group_public_key)|@foreignFieldName groupMessagesReceived|@fieldName receiverAccessGroup';
			comment on table nft_bid_entry is E'@name nft_bid\n@foreignKey (bidder_pkid) references account (pkid)|@foreignFieldName nftBids|@fieldName bidder\n@foreignKey (nft_post_hash) references post_entry (post_hash)|@foreignFieldName nftBids|@fieldName post\n@foreignKey (accepted_block_height) references block (height)|@foreignFieldName nftBids|@fieldName block';
			comment on table nft_entry is E'@name nft\n@foreignKey (last_owner_pkid) references account (pkid)|@foreignFieldName nftsAsLastOwner|@fieldName lastOwner\n@foreignKey (owner_pkid) references account (pkid)|@foreignFieldName nftsOwned|@fieldName owner\n@foreignKey (nft_post_hash) references post_entry (post_hash)|@foreignFieldName nfts|@fieldName post';
			comment on table post_association_entry is E'@name post_association\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName postAssociations|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName postAssociationsAsAppOwner|@fieldName app\n@foreignKey (post_hash) references post_entry (post_hash)|@fieldName post\n@foreignKey (block_height) references block (height)|@foreignFieldName postAssociations|@fieldName block';
			comment on table post_entry is E'@name post\n@foreignKey (poster_public_key) references account (public_key)|@foreignFieldName posts|@fieldName poster\n@foreignKey (parent_post_hash) references post_entry (post_hash)|@foreignFieldName replies|@fieldName parentPost\n@foreignKey (reposted_post_hash) references post_entry (post_hash)|@foreignFieldName reposts|@fieldName repostedPost';
			comment on table profile_entry is E'@name profile\n@foreignKey (public_key) references account (public_key)|@foreignFieldName profile|@fieldName account\n@unique username';
			comment on table transaction is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName transactions|@fieldName block';
			comment on table user_association_entry is E'@name user_association\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTransactor|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName userAssociationsAsAppOwner|@fieldName app\n@foreignKey (target_user_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTarget|@fieldName target\n@foreignKey (block_height) references block (height)|@foreignFieldName userAssociations|@fieldName block';
			comment on table utxo_operation is E'@foreignKey (block_hash, transaction_index) references transaction (block_hash, index_in_block)|@fieldName transaction';
			comment on column access_group_entry.badger_key is E'@omit';
			comment on column access_group_member_entry.badger_key is E'@omit';
			comment on column balance_entry.badger_key is E'@omit';
			comment on column block.badger_key is E'@omit';
			comment on column derived_key_entry.badger_key is E'@omit';
			comment on column deso_balance_entry.badger_key is E'@omit';
			comment on column diamond_entry.badger_key is E'@omit';
			comment on column follow_entry.badger_key is E'@omit';
			comment on column like_entry.badger_key is E'@omit';
			comment on column message_entry.badger_key is E'@omit';
			comment on column new_message_entry.badger_key is E'@omit';
			comment on column nft_bid_entry.badger_key is E'@omit';
			comment on column nft_entry.badger_key is E'@omit';
			comment on column pkid_entry.badger_key is E'@omit';
			comment on column post_association_entry.badger_key is E'@omit';
			comment on column post_entry.badger_key is E'@omit';
			comment on column profile_entry.badger_key is E'@omit';
			comment on column transaction.badger_key is E'@omit';
			comment on column user_association_entry.badger_key is E'@omit';
			comment on table bun_migrations is E'@omit';
			comment on table bun_migration_locks is E'@omit';
			comment on column deso_balance_entry.pkid is E'@name public_key';
			comment on table pkid_entry is E'@omit';
			comment on view wallet is E'@omit';
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
			comment on column access_group_entry.badger_key is NULL;
			comment on column access_group_member_entry.badger_key is NULL;
			comment on column balance_entry.badger_key is NULL;
			comment on column block.badger_key is NULL;
			comment on column derived_key_entry.badger_key is NULL;
			comment on column deso_balance_entry.badger_key is NULL;
			comment on column diamond_entry.badger_key is NULL;
			comment on column follow_entry.badger_key is NULL;
			comment on column like_entry.badger_key is NULL;
			comment on column message_entry.badger_key is NULL;
			comment on column new_message_entry.badger_key is NULL;
			comment on column nft_bid_entry.badger_key is NULL;
			comment on column nft_entry.badger_key is NULL;
			comment on column pkid_entry.badger_key is NULL;
			comment on column post_association_entry.badger_key is NULL;
			comment on column post_entry.badger_key is NULL;
			comment on column profile_entry.badger_key is NULL;
			comment on column transaction.badger_key is NULL;
			comment on column user_association_entry.badger_key is NULL;
			comment on table bun_migrations is NULL;
			comment on table bun_migration_locks is NULL;
			comment on column deso_balance_entry.pkid is NULL';
			comment on table pkid_entry is NULL;
			comment on view wallet is NULL;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
