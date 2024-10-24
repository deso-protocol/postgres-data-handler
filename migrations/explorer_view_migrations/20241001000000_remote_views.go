package explorer_view_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		userPassword := DbPassword
		userName := DbUsername
		dbName := DbName
		host := DbHost
		port := DbPort

		if _, err := db.Exec(`
			CREATE EXTENSION IF NOT EXISTS postgres_fdw;
			
			-- Create a foreign server
			CREATE SERVER IF NOT EXISTS subscriber_server
			FOREIGN DATA WRAPPER postgres_fdw
			OPTIONS (host ?, port ?, dbname ?);
			
			-- Create a user mapping
			CREATE USER MAPPING IF NOT EXISTS FOR current_user
			SERVER subscriber_server
			OPTIONS (user ?, password ?);

			IMPORT FOREIGN SCHEMA public
			LIMIT TO (
				statistic_dashboard,
				statistic_social_leaderboard,
				statistic_nft_leaderboard,
				statistic_defi_leaderboard,
				statistic_txn_count_monthly,
				statistic_wallet_count_monthly,
				statistic_txn_count_daily,
				statistic_new_wallet_count_daily,
				statistic_active_wallet_count_daily,
				statistic_profile_transactions,
				statistic_profile_top_nft_owners,
				dao_coin_limit_order_bid_asks,
				statistic_portfolio_value,
				statistic_profile_earnings,
				statistic_profile_earnings_breakdown_counts
			)
			FROM SERVER subscriber_server
			INTO public;

			CREATE VIEW statistic_dashboard_remote_view AS
			SELECT * FROM statistic_dashboard;
			COMMENT ON VIEW statistic_dashboard_remote_view IS E'@name dashboardStat';
			COMMENT ON FOREIGN TABLE statistic_dashboard IS E'@omit';

			CREATE VIEW statistic_social_leaderboard_remote_view AS
			SELECT * FROM statistic_social_leaderboard;
			COMMENT ON VIEW statistic_social_leaderboard_remote_view IS E'@name socialLeaderboardStat';
			COMMENT ON FOREIGN TABLE statistic_social_leaderboard IS E'@omit';
			
			CREATE VIEW statistic_nft_leaderboard_remote_view AS
			SELECT * FROM statistic_nft_leaderboard;
			COMMENT ON VIEW statistic_nft_leaderboard_remote_view IS E'@name nftLeaderboardStat';
			COMMENT ON FOREIGN TABLE statistic_nft_leaderboard IS E'@omit';
			
			CREATE VIEW statistic_defi_leaderboard_remote_view AS
			SELECT * FROM statistic_defi_leaderboard;
			COMMENT ON VIEW statistic_defi_leaderboard_remote_view IS E'@name defiLeaderboardStat';
			COMMENT ON FOREIGN TABLE statistic_defi_leaderboard IS E'@omit';
			
			CREATE VIEW statistic_txn_count_monthly_remote_view AS
			SELECT * FROM statistic_txn_count_monthly;
			COMMENT ON VIEW statistic_txn_count_monthly_remote_view IS E'@name monthlyTxnCountStat';
			COMMENT ON FOREIGN TABLE statistic_txn_count_monthly IS E'@omit';
			
			CREATE VIEW statistic_wallet_count_monthly_remote_view AS
			SELECT * FROM statistic_wallet_count_monthly;
			COMMENT ON VIEW statistic_wallet_count_monthly_remote_view IS E'@name monthlyNewWalletCountStat';
			COMMENT ON FOREIGN TABLE statistic_wallet_count_monthly IS E'@omit';
			
			CREATE VIEW statistic_txn_count_daily_remote_view AS
			SELECT * FROM statistic_txn_count_daily;
			COMMENT ON VIEW statistic_txn_count_daily_remote_view IS E'@name dailyTxnCountStat';
			COMMENT ON FOREIGN TABLE statistic_txn_count_daily IS E'@omit';
			
			CREATE VIEW statistic_new_wallet_count_daily_remote_view AS
			SELECT * FROM statistic_new_wallet_count_daily;
			COMMENT ON VIEW statistic_new_wallet_count_daily_remote_view IS E'@name dailyNewWalletCountStat';
			COMMENT ON FOREIGN TABLE statistic_new_wallet_count_daily IS E'@omit';
			
			CREATE VIEW statistic_active_wallet_count_daily_remote_view AS
			SELECT * FROM statistic_active_wallet_count_daily;
			COMMENT ON VIEW statistic_active_wallet_count_daily_remote_view IS E'@name dailyActiveWalletCountStat';
			COMMENT ON FOREIGN TABLE statistic_active_wallet_count_daily IS E'@omit';
			
			CREATE VIEW statistic_profile_transactions_remote_view AS
			SELECT * FROM statistic_profile_transactions;
			COMMENT ON VIEW statistic_profile_transactions_remote_view IS E'@name profileTransactionStat\n@unique public_key\n@foreignKey (public_key) references account (public_key)|@foreignFieldName transactionStats|@fieldName account';
			COMMENT ON FOREIGN TABLE statistic_profile_transactions IS E'@omit';
			
			CREATE VIEW statistic_profile_top_nft_owners_remote_view AS
			SELECT * FROM statistic_profile_top_nft_owners;
			COMMENT ON VIEW statistic_profile_top_nft_owners_remote_view IS E'@name profileNftTopOwners';
			COMMENT ON FOREIGN TABLE statistic_profile_top_nft_owners IS E'@omit';
			
			CREATE VIEW dao_coin_limit_order_bid_asks_remote_view AS
			SELECT * FROM dao_coin_limit_order_bid_asks;
			COMMENT ON VIEW dao_coin_limit_order_bid_asks_remote_view IS E'@unique selling_creator_pkid,buying_creator_pkid\n@foreignKey (selling_creator_pkid) references account (pkid)|@foreignFieldName bidAskAsSellingToken|@fieldName sellingTokenAccount\n@foreignKey (buying_creator_pkid) references account (pkid)|@foreignFieldName bidAskAsBuyingToken|@fieldName buyingTokenAccount\n@name deso_token_limit_order_bid_asks';
			COMMENT ON FOREIGN TABLE dao_coin_limit_order_bid_asks IS E'@omit';
			
			CREATE VIEW statistic_portfolio_value_remote_view AS
			SELECT * FROM statistic_portfolio_value;
			COMMENT ON VIEW statistic_portfolio_value_remote_view IS E'@name profilePortfolioValueStat\n@unique public_key\n@omit all';
			COMMENT ON FOREIGN TABLE statistic_portfolio_value IS E'@omit';
			
			CREATE VIEW statistic_profile_earnings_remote_view AS
			SELECT * FROM statistic_profile_earnings;
			COMMENT ON VIEW statistic_profile_earnings_remote_view IS E'@name profileEarningsStats\n@unique public_key\n@omit all';
			COMMENT ON FOREIGN TABLE statistic_profile_earnings IS E'@omit';
			
			CREATE VIEW statistic_profile_earnings_breakdown_counts_remote_view AS
			SELECT * FROM statistic_profile_earnings_breakdown_counts;
			COMMENT ON VIEW statistic_profile_earnings_breakdown_counts_remote_view IS E'@name profileEarningsBreakdownStats\n@unique public_key\n@omit all';
			COMMENT ON FOREIGN TABLE statistic_profile_earnings_breakdown_counts IS E'@omit';
		`, host, port, dbName, userName, userPassword); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
		drop index if exists statistic_profile_transactions_latest_idx;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}
