package post_sync_migrations

import (
	"context"
	"errors"
	"fmt"
	"github.com/uptrace/bun"
	"math"
	"time"
)

const (
	retryLimit = 10
)

var (
	commands = []struct {
		Query  string
		Ticker *time.Ticker
	}{
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_all", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_30_d", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_wallet_count_all", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_new_wallet_count_30_d", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_active_wallet_count_30_d", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_block_height_current", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_pending", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_fee_1_d", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_total_supply", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_post_count", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_comment_count", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_repost_count", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_creator_coin", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_nft", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_dex", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_social", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_follow_count", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_message_count", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "SELECT refresh_public_key_first_transaction()", Ticker: time.NewTicker(10 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_likes", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_reactions", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_diamonds", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_reposts", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_comments", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard", Ticker: time.NewTicker(1 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_nft_leaderboard", Ticker: time.NewTicker(1 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_defi_leaderboard", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_monthly", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_wallet_count_monthly", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_daily", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_new_wallet_count_daily", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_active_wallet_count_daily", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_transactions", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_top_nft_owners", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_cc_balance_totals", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_nft_balance_totals", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_deso_token_balance_totals", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_portfolio_value", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_royalties", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamond_earnings", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_royalty_earnings", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_royalty_earnings", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_buy_orders", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_sell_orders", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamonds_given", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamonds_received", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_buyers", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_sellers", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_buys", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_sales", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_buys", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_sales", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_buy_orders", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_sell_orders", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_earnings_breakdown_counts", Ticker: time.NewTicker(30 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY staking_summary", Ticker: time.NewTicker(1 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY my_stake_summary", Ticker: time.NewTicker(1 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY validator_stats", Ticker: time.NewTicker(1 * time.Second)},
	}
)

func RunMigrationWithRetries(db *bun.DB, migrationQuery string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()
	for ii := 0; ii < retryLimit; ii++ {
		_, err := db.ExecContext(ctx, migrationQuery)
		if err == nil {
			return nil
		}
		waitTime := 5 * time.Duration(math.Pow(2, float64(ii))) * time.Second
		fmt.Printf("Failed to migrate, retrying in %v. err: %v. Query: %v\n", waitTime, err, migrationQuery)
		time.Sleep(waitTime)
	}
	return errors.New("Failed to migrate after 5 attempts")
}

func RefreshExplorerStatistics(db *bun.DB) {
	// Only run if explorer statistics are enabled.
	if !calculateExplorerStatistics {
		return
	}

	// Run each refresh command in a non-blocking goroutine.
	for _, command := range commands {
		go func(command struct {
			Query  string
			Ticker *time.Ticker
		}) {
			// Create a channel to ensure only one command is running at a time.
			running := make(chan bool, 1)
			for range command.Ticker.C {
				// If a command is still running, skip
				if len(running) > 0 {
					continue
				}

				running <- true
				go func() {
					err := executeQuery(db, command.Query)
					if err != nil {
						fmt.Printf("Error executing explorer refresh query: %s: %v\n", command.Query, err)
					}
					<-running
				}()
			}
		}(command)
	}

	// Wait indefinitely.
	select {}
}
