package post_sync_migrations

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/uptrace/bun"
	"math"
	"time"
)

const (
	retryLimit = 10
)

var (
	commands = []struct {
		Name     string
		Query    string
		Duration time.Duration
	}{
		{Name: "refresh_public_key_first_transaction", Query: "SELECT refresh_public_key_first_transaction()", Duration: 1 * time.Hour},
		{Name: "statistic_txn_count_all", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_all", Duration: 15 * time.Minute},
		{Name: "statistic_txn_count_30_d", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_30_d", Duration: 30 * time.Minute},
		{Name: "statistic_wallet_count_all", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_wallet_count_all", Duration: 15 * time.Minute},
		{Name: "statistic_new_wallet_count_30_d", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_new_wallet_count_30_d", Duration: 15 * time.Minute},
		{Name: "statistic_active_wallet_count_30_d", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_active_wallet_count_30_d", Duration: 2 * time.Hour},
		{Name: "statistic_block_height_current", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_block_height_current", Duration: 2 * time.Second},
		{Name: "statistic_txn_count_pending", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_pending", Duration: 15 * time.Minute},
		{Name: "statistic_txn_fee_1_d", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_fee_1_d", Duration: 15 * time.Minute},
		{Name: "statistic_total_supply", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_total_supply", Duration: 15 * time.Minute},
		{Name: "statistic_post_count", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_post_count", Duration: 15 * time.Minute},
		{Name: "statistic_comment_count", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_comment_count", Duration: 15 * time.Minute},
		{Name: "statistic_repost_count", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_repost_count", Duration: 15 * time.Minute},
		{Name: "statistic_txn_count_creator_coin", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_creator_coin", Duration: 15 * time.Minute},
		{Name: "statistic_txn_count_nft", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_nft", Duration: 15 * time.Minute},
		{Name: "statistic_txn_count_dex", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_dex", Duration: 15 * time.Minute},
		{Name: "statistic_txn_count_social", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_social", Duration: 15 * time.Minute},
		{Name: "statistic_follow_count", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_follow_count", Duration: 15 * time.Minute},
		{Name: "statistic_message_count", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_message_count", Duration: 15 * time.Minute},
		{Name: "statistic_social_leaderboard_likes", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_likes", Duration: 30 * time.Minute},
		{Name: "statistic_social_leaderboard_reactions", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_reactions", Duration: 15 * time.Minute},
		{Name: "statistic_social_leaderboard_diamonds", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_diamonds", Duration: 15 * time.Minute},
		{Name: "statistic_social_leaderboard_reposts", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_reposts", Duration: 15 * time.Minute},
		{Name: "statistic_social_leaderboard_comments", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard_comments", Duration: 15 * time.Minute},
		{Name: "statistic_social_leaderboard", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_social_leaderboard", Duration: 1 * time.Minute},
		{Name: "statistic_nft_leaderboard", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_nft_leaderboard", Duration: 1 * time.Minute},
		{Name: "statistic_defi_leaderboard", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_defi_leaderboard", Duration: 30 * time.Minute},
		{Name: "statistic_txn_count_monthly", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_monthly", Duration: 30 * time.Minute},
		{Name: "statistic_wallet_count_monthly", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_wallet_count_monthly", Duration: 30 * time.Minute},
		{Name: "statistic_txn_count_daily", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_txn_count_daily", Duration: 30 * time.Minute},
		{Name: "statistic_new_wallet_count_daily", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_new_wallet_count_daily", Duration: 30 * time.Minute},
		{Name: "statistic_active_wallet_count_daily", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_active_wallet_count_daily", Duration: 30 * time.Minute},
		{Name: "statistic_profile_transactions", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_transactions", Duration: 1 * time.Hour},
		{Name: "statistic_profile_top_nft_owners", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_top_nft_owners", Duration: 30 * time.Minute},
		{Name: "statistic_cc_balance_totals", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_cc_balance_totals", Duration: 30 * time.Minute},
		{Name: "statistic_nft_balance_totals", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_nft_balance_totals", Duration: 30 * time.Minute},
		{Name: "statistic_deso_token_balance_totals", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_deso_token_balance_totals", Duration: 30 * time.Minute},
		{Name: "statistic_portfolio_value", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_portfolio_value", Duration: 3 * time.Hour},
		{Name: "statistic_profile_cc_royalties", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_royalties", Duration: 30 * time.Minute},
		{Name: "statistic_profile_diamond_earnings", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamond_earnings", Duration: 30 * time.Minute},
		{Name: "statistic_profile_nft_bid_royalty_earnings", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_royalty_earnings", Duration: 30 * time.Minute},
		{Name: "statistic_profile_nft_buy_now_royalty_earnings", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_royalty_earnings", Duration: 30 * time.Minute},
		{Name: "statistic_profile_deso_token_buy_orders", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_buy_orders", Duration: 30 * time.Minute},
		{Name: "statistic_profile_deso_token_sell_orders", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_sell_orders", Duration: 30 * time.Minute},
		{Name: "statistic_profile_diamonds_given", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamonds_given", Duration: 30 * time.Minute},
		{Name: "statistic_profile_diamonds_received", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_diamonds_received", Duration: 30 * time.Minute},
		{Name: "statistic_profile_cc_buyers", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_buyers", Duration: 3 * time.Hour},
		{Name: "statistic_profile_cc_sellers", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_cc_sellers", Duration: 3 * time.Hour},
		{Name: "statistic_profile_nft_bid_buys", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_buys", Duration: 1 * time.Hour},
		{Name: "statistic_profile_nft_bid_sales", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_bid_sales", Duration: 30 * time.Minute},
		{Name: "statistic_profile_nft_buy_now_buys", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_buys", Duration: 30 * time.Minute},
		{Name: "statistic_profile_nft_buy_now_sales", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_nft_buy_now_sales", Duration: 30 * time.Minute},
		{Name: "statistic_profile_deso_token_buy_orders", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_buy_orders", Duration: 30 * time.Minute},
		{Name: "statistic_profile_deso_token_sell_orders", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_deso_token_sell_orders", Duration: 30 * time.Minute},
		{Name: "statistic_profile_earnings_breakdown_counts", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY statistic_profile_earnings_breakdown_counts", Duration: 30 * time.Minute},
		{Name: "staking_summary", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY staking_summary", Duration: 1 * time.Minute},
		{Name: "my_stake_summary", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY my_stake_summary", Duration: 1 * time.Minute},
		{Name: "validator_stats", Query: "REFRESH MATERIALIZED VIEW CONCURRENTLY validator_stats", Duration: 1 * time.Minute},
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
	err := StartAsyncJobs(db)
	if err != nil {
		fmt.Printf("Error starting async jobs: %v\n", err)
	}

	// Wait indefinitely.
	select {}
}

var AsyncJobScheduler gocron.Scheduler

type AsyncJob struct {
	Name     string
	Schedule gocron.JobDefinition
	Task     gocron.Task
}

func StartAsyncJobs(db *bun.DB) error {
	var asyncJobs []*AsyncJob

	for _, command := range commands {
		asyncJobs = append(asyncJobs, &AsyncJob{
			Name:     command.Name,
			Schedule: gocron.DurationJob(command.Duration),
			Task: gocron.NewTask(func() {
				if err := executeQuery(db, command.Query); err != nil {
					fmt.Printf("Error executing explorer refresh query: %s: %v\n", command.Query, err)
				}
			}),
		})
	}

	var err error
	AsyncJobScheduler, err = gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return fmt.Errorf("error starting async job scheduler: %v", err)
	}

	for _, job := range asyncJobs {
		if _, err = AsyncJobScheduler.NewJob(
			job.Schedule,
			job.Task,
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		); err != nil {
			return fmt.Errorf("error scheduling async job %s: %v", job.Name, err)
		}
	}

	AsyncJobScheduler.Start()
	return nil
}

func StopAsyncJobs() {
	if AsyncJobScheduler != nil {
		AsyncJobScheduler.Shutdown()
	}
}
