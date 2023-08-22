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
			comment on table post_association_entry is E'@name post_association\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName postAssociations|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName postAssociationsAsAppOwner|@fieldName app\n@foreignKey (post_hash) references post_entry (post_hash)|@fieldName post\n@foreignKey (block_height) references block (height)|@foreignFieldName postAssociations|@fieldName block';
			comment on table post_entry is E'@name post\n@foreignKey (poster_public_key) references account (public_key)|@foreignFieldName posts|@fieldName poster\n@foreignKey (parent_post_hash) references post_entry (post_hash)|@foreignFieldName replies|@fieldName parentPost\n@foreignKey (reposted_post_hash) references post_entry (post_hash)|@foreignFieldName reposts|@fieldName repostedPost';
			comment on table profile_entry is E'@name profile\n@foreignKey (public_key) references account (public_key)|@foreignFieldName profile|@fieldName account\n@unique username';
			comment on view transaction is E'@foreignKey (block_hash) references block (block_hash)|@foreignFieldName transactions|@fieldName block\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactions|@fieldName account\n@unique transaction_hash';
			comment on table transaction_type is E'@foreignKey (type) references transaction (txn_type)|@foreignFieldName transactionType|@fieldName transaction';
			comment on table user_association_entry is E'@name user_association\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTransactor|@fieldName transactor\n@foreignKey (app_pkid) references account (pkid)|@foreignFieldName userAssociationsAsAppOwner|@fieldName app\n@foreignKey (target_user_pkid) references account (pkid)|@foreignFieldName userAssociationsAsTarget|@fieldName target\n@foreignKey (block_height) references block (height)|@foreignFieldName userAssociations|@fieldName block';
			comment on table utxo_operation is E'@foreignKey (block_hash, transaction_index) references transaction (block_hash, index_in_block)|@fieldName transaction';
			comment on table dao_coin_limit_order_entry is E'@name desoTokenLimitOrder\n@foreignKey (transactor_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByTransactor|@fieldName transactorAccount\n@foreignKey (buying_dao_coin_creator_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByCreatorBought|@fieldName creatorBoughtAccount\n@foreignKey (selling_dao_coin_creator_pkid) references account (pkid)|@foreignFieldName desoTokenLimitOrderByCreatorSold|@fieldName creatorSoldAccount\n@unique order_id\n@foreignKey (selling_dao_coin_creator_pkid, transactor_pkid) references balance_entry (creator_pkid, hodler_pkid)|@foreignFieldName desoTokenLimitOrders|@fieldName transactorSellingTokenBalance';
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

		// Only annotate the explorer statistics views if the env var is set to enable them.
		if !calculateExplorerStatistics {
			return nil
		}

		_, err = db.Exec(`
			comment on materialized view statistic_txn_count_all is E'@omit';
			comment on materialized view statistic_txn_count_30_d is E'@omit';
			comment on materialized view statistic_wallet_count_all is E'@omit';
			comment on materialized view statistic_new_wallet_count_30_d is E'@omit';
			comment on materialized view statistic_active_wallet_count_30_d is E'@omit';
			comment on materialized view statistic_block_height_current is E'@omit';
			comment on materialized view statistic_txn_count_pending is E'@omit';
			comment on materialized view statistic_txn_fee_1_d is E'@omit';
			comment on materialized view statistic_total_supply is E'@omit';
			comment on materialized view statistic_post_count is E'@omit';
			comment on materialized view statistic_post_longform_count is E'@omit';
			comment on materialized view statistic_comment_count is E'@omit';
			comment on materialized view statistic_repost_count is E'@omit';
			comment on materialized view statistic_txn_count_creator_coin is E'@omit';
			comment on materialized view statistic_txn_count_nft is E'@omit';
			comment on materialized view statistic_txn_count_dex is E'@omit';
			comment on materialized view statistic_txn_count_social is E'@omit';
			comment on materialized view statistic_follow_count is E'@omit';
			comment on materialized view statistic_message_count is E'@omit';
			comment on materialized view statistic_social_leaderboard_likes is E'@omit';
			comment on materialized view statistic_social_leaderboard_reactions is E'@omit';
			comment on materialized view statistic_social_leaderboard_diamonds is E'@omit';
			comment on materialized view statistic_social_leaderboard_reposts is E'@omit';
			comment on materialized view statistic_social_leaderboard_comments is E'@omit';
			comment on table public_key_first_transaction IS E'@omit';
			comment on function get_transaction_count is E'@omit';
			comment on function hex_to_decimal is E'@omit';
			comment on function refresh_public_key_first_transaction is E'@omit';
			comment on view statistic_dashboard is E'@name dashboardStat';
			comment on materialized view statistic_social_leaderboard is E'@name socialLeaderboardStat';
			comment on materialized view statistic_nft_leaderboard is E'@name nftLeaderboardStat';
			comment on materialized view statistic_defi_leaderboard is E'@name defiLeaderboardStat';
			comment on materialized view statistic_txn_count_monthly is E'@name monthlyTxnCountStat';
			comment on materialized view statistic_wallet_count_monthly is E'@name monthlyNewWalletCountStat';
			comment on materialized view statistic_txn_count_daily is E'@name dailyTxnCountStat';
			comment on materialized view statistic_new_wallet_count_daily is E'@name dailyNewWalletCountStat';
			comment on materialized view statistic_active_wallet_count_daily is E'@name dailyActiveWalletCountStat';
			comment on materialized view statistic_profile_transactions is E'@name profileTransactionStat\n@unique public_key\n@omit all';
			comment on materialized view statistic_profile_top_nft_owners is E'@name profileNftTopOwners';
			comment on function hex_to_numeric is E'@omit';
			comment on function cc_nanos_total_sell_value is E'@omit';
			comment on view dao_coin_limit_order_max_bids is E'@omit';
			comment on view dao_coin_limit_order_min_asks is E'@omit';
			comment on view dao_coin_limit_order_bid_asks is E'@omit';
			comment on materialized view statistic_cc_balance_totals is E'@omit';
			comment on materialized view statistic_nft_balance_totals is E'@omit';
			comment on materialized view statistic_deso_token_balance_totals is E'@omit';
			comment on materialized view statistic_portfolio_value is E'@name profilePortfolioValueStat\n@unique public_key\n@omit all';
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

		// Only revert the explorer statistics views if the env var is set to enable them.
		if !calculateExplorerStatistics {
			return nil
		}

		_, err = db.Exec(`
			comment on materialized view statistic_txn_count_all is NULL;
			comment on materialized view statistic_txn_count_30_d is NULL;
			comment on materialized view statistic_wallet_count_all is NULL;
			comment on materialized view statistic_new_wallet_count_30_d is NULL;
			comment on materialized view statistic_active_wallet_count_30_d is NULL;
			comment on materialized view statistic_block_height_current is NULL;
			comment on materialized view statistic_txn_count_pending is NULL;
			comment on materialized view statistic_txn_fee_1_d is NULL;
			comment on materialized view statistic_total_supply is NULL;
			comment on materialized view statistic_post_count is NULL;
			comment on materialized view statistic_post_longform_count is NULL;
			comment on materialized view statistic_comment_count is NULL;
			comment on materialized view statistic_repost_count is NULL;
			comment on materialized view statistic_txn_count_creator_coin is NULL;
			comment on materialized view statistic_txn_count_nft is NULL;
			comment on materialized view statistic_txn_count_dex is NULL;
			comment on materialized view statistic_txn_count_social is NULL;
			comment on materialized view statistic_follow_count is NULL;
			comment on materialized view statistic_message_count is NULL;
			comment on materialized view statistic_social_leaderboard_likes is NULL;
			comment on materialized view statistic_social_leaderboard_reactions is NULL;
			comment on materialized view statistic_social_leaderboard_diamonds is NULL;
			comment on materialized view statistic_social_leaderboard_reposts is NULL;
			comment on materialized view statistic_social_leaderboard_comments is NULL;
			comment on table public_key_first_transaction IS NULL;
			comment on function get_transaction_count is NULL;
			comment on function hex_to_decimal is NULL;
			comment on function refresh_public_key_first_transaction is NULL;
			comment on view statistic_dashboard is NULL;
			comment on materialized view statistic_social_leaderboard is NULL;
			comment on materialized view statistic_nft_leaderboard is NULL;
			comment on materialized view statistic_defi_leaderboard is NULL;
			comment on materialized view statistic_txn_count_monthly is NULL;
			comment on materialized view statistic_wallet_count_monthly is NULL;
			comment on materialized view statistic_wallet_count_monthly is NULL;
			comment on materialized view statistic_txn_count_daily is NULL;
			comment on materialized view statistic_new_wallet_count_daily is NULL;
			comment on materialized view statistic_active_wallet_count_daily is NULL;
			comment on materialized view statistic_profile_transactions is NULL;
			comment on materialized view statistic_profile_top_nft_owners is NULL;
			comment on function cc_nanos_total_sell_value is NULL;
			comment on view dao_coin_limit_order_max_bids is NULL;
			comment on view dao_coin_limit_order_min_asks is NULL;
			comment on view dao_coin_limit_order_bid_asks is NULL;
			comment on materialized view statistic_cc_balance_totals is NULL;
			comment on materialized view statistic_nft_balance_totals is NULL;
			comment on materialized view statistic_deso_token_balance_totals is NULL;
			comment on materialized view statistic_portfolio_value is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
