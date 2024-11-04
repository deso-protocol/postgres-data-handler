package explorer_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
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
			comment on materialized view statistic_profile_transactions is E'@name profileTransactionStat\n@unique public_key\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactionStats|@fieldName account';
			comment on materialized view statistic_profile_top_nft_owners is E'@name profileNftTopOwners';
			comment on function hex_to_numeric is E'@omit';
			comment on function int_to_bytea is E'@omit';
			comment on function cc_nanos_total_sell_value is E'@omit';
			comment on view dao_coin_limit_order_max_bids is E'@omit';
			comment on view dao_coin_limit_order_min_asks is E'@omit';
			comment on view dao_coin_limit_order_bid_asks is E'@unique selling_creator_pkid,buying_creator_pkid\n@foreignKey (selling_creator_pkid) references account (pkid)|@foreignFieldName bidAskAsSellingToken|@fieldName sellingTokenAccount\n@foreignKey (buying_creator_pkid) references account (pkid)|@foreignFieldName bidAskAsBuyingToken|@fieldName buyingTokenAccount\n@name deso_token_limit_order_bid_asks';
			comment on materialized view statistic_cc_balance_totals is E'@omit';
			comment on materialized view statistic_nft_balance_totals is E'@omit';
			comment on materialized view statistic_deso_token_balance_totals is E'@omit';
			comment on materialized view statistic_portfolio_value is E'@name profilePortfolioValueStat\n@unique public_key\n@omit all';
			comment on materialized view statistic_profile_cc_royalties is E'@omit';
			comment on materialized view statistic_profile_diamond_earnings is E'@omit';
			comment on materialized view statistic_profile_nft_bid_royalty_earnings is E'@omit';
			comment on materialized view statistic_profile_nft_buy_now_royalty_earnings is E'@omit';
			comment on materialized view statistic_profile_earnings is E'@name profileEarningsStats\n@unique public_key\n@omit all';
			comment on materialized view statistic_profile_deso_token_buy_orders is E'@omit';
			comment on materialized view statistic_profile_deso_token_sell_orders is E'@omit';
			comment on materialized view statistic_profile_diamonds_given is E'@omit';
			comment on materialized view statistic_profile_diamonds_received is E'@omit';
			comment on materialized view statistic_profile_cc_buyers is E'@omit';
			comment on materialized view statistic_profile_cc_sellers is E'@omit';
			comment on materialized view statistic_profile_nft_bid_buys is E'@omit';
			comment on materialized view statistic_profile_nft_bid_sales is E'@omit';
			comment on materialized view statistic_profile_nft_buy_now_buys is E'@omit';
			comment on materialized view statistic_profile_nft_buy_now_sales is E'@omit';
			comment on materialized view statistic_profile_deso_token_buy_orders is E'@omit';
			comment on materialized view statistic_profile_deso_token_sell_orders is E'@omit';
			comment on materialized view statistic_profile_earnings_breakdown_counts is E'@name profileEarningsBreakdownStats\n@unique public_key\n@omit all';
			comment on function jsonb_to_bytea is E'@omit';
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
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
			comment on materialized view statistic_profile_deso_token_buy_orders is NULL;
			comment on materialized view statistic_profile_deso_token_sell_orders is NULL;
			comment on materialized view statistic_profile_diamonds_given is NULL;
			comment on materialized view statistic_profile_diamonds_received is NULL;
			comment on materialized view statistic_profile_cc_buyers is NULL;
			comment on materialized view statistic_profile_cc_sellers is NULL;
			comment on materialized view statistic_profile_nft_bid_buys is NULL;
			comment on materialized view statistic_profile_nft_bid_sales is NULL;
			comment on materialized view statistic_profile_nft_buy_now_buys is NULL;
			comment on materialized view statistic_profile_nft_buy_now_sales is NULL;
			comment on materialized view statistic_profile_deso_token_buy_orders is NULL;
			comment on materialized view statistic_profile_deso_token_sell_orders is NULL;
			comment on materialized view statistic_profile_earnings_breakdown_counts is NULL;
			comment on function jsonb_to_bytea is NULL;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
