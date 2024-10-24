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
			comment on table access_group_entry is E'@name access_group\n@unique access_group_owner_public_key,access_group_key_name\n@unique access_group_public_key\n@foreignKey (access_group_owner_public_key) references account (public_key)|@foreignFieldName accessGroupsOwned|@fieldName owner';
			comment on table access_group_member_entry is E'@name access_group_member\n@unique access_group_owner_public_key, access_group_member_public_key, access_group_key_name, access_group_member_key_name\n@foreignKey (access_group_member_public_key) references account (public_key)|@foreignFieldName accessGroupMemberships|@fieldName member\n@foreignKey (access_group_owner_public_key, access_group_key_name) references access_group_entry (access_group_owner_public_key, access_group_key_name)|@foreignFieldName accessGroupMembers|@fieldName accessGroup';
			comment on table affected_public_key is E'@foreignKey (public_key) references account (public_key)|@foreignFieldName transactionHashes|@fieldName account\n@foreignKey (transaction_hash, txn_type) references transaction (transaction_hash, txn_type)|@fieldName transaction|@foreignFieldName affectedPublicKeys';
			comment on table balance_entry is E'@name tokenBalance\n@foreignKey (hodler_pkid) references account (pkid)|@foreignFieldName tokenBalances|@fieldName holder\n@foreignKey (creator_pkid) references account (pkid)|@foreignFieldName tokenBalancesAsCreator|@fieldName creator';
			comment on view creator_coin_balance is E'@primaryKey hodler_pkid,creator_pkid\n@name creatorCoinBalance\n@foreignKey (hodler_pkid) references account (pkid)|@foreignFieldName creatorCoinBalances|@fieldName holder\n@foreignKey (creator_pkid) references account (pkid)|@foreignFieldName creatorCoinBalancesAsCreator|@fieldName creator';
			comment on table derived_key_entry is E'@name derived_key\n@foreignKey (owner_public_key) references account (public_key)|@foreignFieldName derivedKeys|@fieldName owner';
			comment on table deso_balance_entry is E'@name desoBalance\n@unique public_key\n@primaryKey public_key\n@foreignKey (public_key) references account (public_key)|@fieldName account|@foreignFieldName desoBalance';
			comment on table diamond_entry is E'@name diamond\n@foreignKey (sender_pkid) references account (pkid)|@foreignFieldName diamondsSent|@fieldName sender\n@foreignKey (receiver_pkid) references account (pkid)|@foreignFieldName diamondsReceived|@fieldName reciever\n@foreignKey (post_hash) references post_entry (post_hash)|@foreignFieldName diamonds|@fieldName post';
			comment on table follow_entry is E'@name follow\n@foreignKey (follower_pkid) references account (pkid)|@foreignFieldName following|@fieldName follower\n@foreignKey (followed_pkid) references account (pkid)|@foreignFieldName followers|@fieldName followee';
			comment on table like_entry is E'@name like\n@foreignKey (public_key) references account (public_key)|@foreignFieldName likes|@fieldName account\n@foreignKey (post_hash) references post_entry (post_hash)|@foreignFieldName likes|@fieldName post';
			comment on table message_entry is E'@name legacyMessage\n@foreignKey (sender_public_key) references account (public_key)|@foreignFieldName legacyMessagesSent|@fieldName sender\n@foreignKey (recipient_public_key) references account (public_key)|@foreignFieldName legacyMessagesReceived|@fieldName receiver';
			comment on table new_message_entry is E'@name message\n@foreignKey (sender_access_group_owner_public_key) references account (public_key)|@foreignFieldName messagesSent|@fieldName sender\n@foreignKey (recipient_access_group_owner_public_key) references account (public_key)|@foreignFieldName messagesReceived|@fieldName receiver\n@foreignKey (sender_access_group_public_key) references access_group_entry (access_group_public_key)|@foreignFieldName groupMessagesSent|@fieldName senderAccessGroup\n@foreignKey (recipient_access_group_public_key) references access_group_entry (access_group_public_key)|@foreignFieldName groupMessagesReceived|@fieldName receiverAccessGroup';
			comment on table nft_bid_entry is E'@name nft_bid\n@foreignKey (bidder_pkid) references account (pkid)|@foreignFieldName nftBids|@fieldName bidder\n@foreignKey (nft_post_hash) references post_entry (post_hash)|@foreignFieldName nftBids|@fieldName post\n@foreignKey (accepted_block_height) references block (height)|@foreignFieldName nftBids|@fieldName block';
			comment on table nft_entry is E'@name nft\n@foreignKey (last_owner_pkid) references account (pkid)|@foreignFieldName nftsAsLastOwner|@fieldName lastOwner\n@foreignKey (owner_pkid) references account (pkid)|@foreignFieldName nftsOwned|@fieldName owner\n@foreignKey (nft_post_hash) references post_entry (post_hash)|@foreignFieldName nfts|@fieldName post';
			comment on table post_association_entry is E'@name post_association\n@unique association_id\n@unique transactor_pkid, post_hash, app_pkid, association_type, association_value\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName postAssociations|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName postAssociationsAsAppOwner|@fieldName app\n@foreignKey (post_hash) references post_entry (post_hash)|@fieldName post\n@foreignKey (block_height) references block (height)|@foreignFieldName postAssociations|@fieldName block';
			comment on table post_entry is E'@name post\n@unique post_hash\n@foreignKey (poster_public_key) references account (public_key)|@foreignFieldName posts|@fieldName poster\n@foreignKey (parent_post_hash) references post_entry (post_hash)|@foreignFieldName replies|@fieldName parentPost\n@foreignKey (reposted_post_hash) references post_entry (post_hash)|@foreignFieldName reposts|@fieldName repostedPost';
			comment on table profile_entry is E'@name profile\n@foreignKey (public_key) references account (public_key)|@foreignFieldName profile|@fieldName account\n@unique username';
			comment on view transaction is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName transactions|@fieldName block\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactions|@fieldName account\n@unique transaction_hash';
			comment on table transaction_type is E'@foreignKey (type) references transaction (txn_type)|@foreignFieldName transactionType|@fieldName transaction';
			comment on table user_association_entry is E'@name user_association\n@unique association_id\n@unique transactor_pkid, target_user_pkid, app_pkid, association_type, association_value\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTransactor|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName userAssociationsAsAppOwner|@fieldName app\n@foreignKey (target_user_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTarget|@fieldName target\n@foreignKey (block_height) references block (height)|@foreignFieldName userAssociations|@fieldName block';
			comment on table utxo_operation is E'@foreignKey (block_hash, transaction_index) references transaction (block_hash, index_in_block)|@fieldName transaction';
			comment on table dao_coin_limit_order_entry is E'@name desoTokenLimitOrder\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByTransactor|@fieldName transactorAccount\n@foreignKey (buying_dao_coin_creator_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByCreatorBought|@fieldName creatorBoughtAccount\n@foreignKey (selling_dao_coin_creator_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByCreatorSold|@fieldName creatorSoldAccount\n@unique order_id\n@foreignKey (selling_dao_coin_creator_pkid, transactor_pkid, is_dao_coin_const) references balance_entry (creator_pkid, hodler_pkid, is_dao_coin)|@foreignFieldName desoTokenSellingLimitOrders|@fieldName transactorSellingTokenBalance\n@foreignKey (buying_dao_coin_creator_pkid, transactor_pkid, is_dao_coin_const) references balance_entry (creator_pkid, hodler_pkid, is_dao_coin)|@foreignFieldName desoTokenBuyingLimitOrders|@fieldName transactorBuyingTokenBalance';
			comment on table block is E'@unique block_hash\n@unique height';
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
			comment on column account.token_balance_join_field is E'@omit';
			comment on column account.cc_balance_join_field is E'@omit';
			comment on column user_association_entry.badger_key is E'@omit';
			comment on table bun_migrations is E'@omit';
			comment on table bun_migration_locks is E'@omit';
			comment on table pkid_entry is E'@omit';
			comment on table transaction_partitioned is E'@omit';
			comment on table transaction_partition_01 is E'@omit';
			comment on table transaction_partition_02 is E'@omit';
			comment on table transaction_partition_03 is E'@omit';
			comment on table transaction_partition_04 is E'@omit';
			comment on table transaction_partition_05 is E'@omit';
			comment on table transaction_partition_06 is E'@omit';
			comment on table transaction_partition_07 is E'@omit';
			comment on table transaction_partition_08 is E'@omit';
			comment on table transaction_partition_09 is E'@omit';
			comment on table transaction_partition_10 is E'@omit';
			comment on table transaction_partition_11 is E'@omit';
			comment on table transaction_partition_12 is E'@omit';
			comment on table transaction_partition_13 is E'@omit';
			comment on table transaction_partition_14 is E'@omit';
			comment on table transaction_partition_15 is E'@omit';
			comment on table transaction_partition_16 is E'@omit';
			comment on table transaction_partition_17 is E'@omit';
			comment on table transaction_partition_18 is E'@omit';
			comment on table transaction_partition_19 is E'@omit';
			comment on table transaction_partition_20 is E'@omit';
			comment on table transaction_partition_21 is E'@omit';
			comment on table transaction_partition_22 is E'@omit';
			comment on table transaction_partition_23 is E'@omit';
			comment on table transaction_partition_24 is E'@omit';
			comment on table transaction_partition_25 is E'@omit';
			comment on table transaction_partition_26 is E'@omit';
			comment on table transaction_partition_27 is E'@omit';
			comment on table transaction_partition_28 is E'@omit';
			comment on table transaction_partition_29 is E'@omit';
			comment on table transaction_partition_30 is E'@omit';
			comment on table transaction_partition_31 is E'@omit';
			comment on table transaction_partition_32 is E'@omit';
			comment on table transaction_partition_33 is E'@omit';
			comment on function checksum is E'@omit';
			comment on function base58_check_encode_with_prefix is E'@omit';
			comment on function bytes_to_bigint is E'@omit';
			comment on function base58_encode is E'@omit';
			comment on function base64_to_base58 is E'@omit';
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
			comment on view transaction is NULL;
			comment on table user_association_entry is NULL;
			comment on table utxo_operation is NULL;
			comment on column account.token_balance_join_field is NULL;
			comment on column account.cc_balance_join_field is NULL;
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
			comment on table pkid_entry is NULL;
			comment on view wallet is NULL;
			comment on table dao_coin_limit_order_entry is NULL;
			comment on table transaction_partitioned is NULL;
			comment on table transaction_partition_01 is NULL;
			comment on table transaction_partition_02 is NULL;
			comment on table transaction_partition_03 is NULL;
			comment on table transaction_partition_04 is NULL;
			comment on table transaction_partition_05 is NULL;
			comment on table transaction_partition_06 is NULL;
			comment on table transaction_partition_07 is NULL;
			comment on table transaction_partition_08 is NULL;
			comment on table transaction_partition_09 is NULL;
			comment on table transaction_partition_10 is NULL;
			comment on table transaction_partition_11 is NULL;
			comment on table transaction_partition_12 is NULL;
			comment on table transaction_partition_13 is NULL;
			comment on table transaction_partition_14 is NULL;
			comment on table transaction_partition_15 is NULL;
			comment on table transaction_partition_16 is NULL;
			comment on table transaction_partition_17 is NULL;
			comment on table transaction_partition_18 is NULL;
			comment on table transaction_partition_19 is NULL;
			comment on table transaction_partition_20 is NULL;
			comment on table transaction_partition_21 is NULL;
			comment on table transaction_partition_22 is NULL;
			comment on table transaction_partition_23 is NULL;
			comment on table transaction_partition_24 is NULL;
			comment on table transaction_partition_25 is NULL;
			comment on table transaction_partition_26 is NULL;
			comment on table transaction_partition_27 is NULL;
			comment on table transaction_partition_28 is NULL;
			comment on table transaction_partition_09 is NULL;
			comment on table transaction_partition_30 is NULL;
			comment on table transaction_partition_31 is NULL;
			comment on table transaction_partition_32 is NULL;
			comment on table transaction_partition_33 is NULL;
			comment on function checksum is NULL;
			comment on function base58_check_encode_with_prefix is NULL;
			comment on function bytes_to_bigint is NULL;
			comment on function base58_encode is NULL;
			comment on function base64_to_base58 is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
