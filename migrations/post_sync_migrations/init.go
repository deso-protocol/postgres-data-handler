package post_sync_migrations

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"time"
)

var (
	calculateExplorerStatistics bool
	Migrations                  = migrate.NewMigrations()
	commands                    = []struct {
		Query  string
		Ticker *time.Ticker
	}{
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_all", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_30_d", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_wallet_count_all", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_wallet_count_30_d", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_block_height_current", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_pending", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_fee_1_d", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_total_supply", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_post_count", Ticker: time.NewTicker(30 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_comment_count", Ticker: time.NewTicker(30 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_repost_count", Ticker: time.NewTicker(30 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_creator_coin", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_nft", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_dex", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_social", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_follow_count", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_message_count", Ticker: time.NewTicker(60 * time.Second)},
		{Query: "PERFORM refresh_public_key_first_transaction()", Ticker: time.NewTicker(5 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard_likes", Ticker: time.NewTicker(10 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard_reactions", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard_diamonds", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard_reposts", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard_comments", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_social_leaderboard", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_nft_leaderboard", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_defi_leaderboard", Ticker: time.NewTicker(15 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_txn_count_monthly", Ticker: time.NewTicker(10 * time.Minute)},
		{Query: "REFRESH MATERIALIZED VIEW statistic_wallet_count_monthly", Ticker: time.NewTicker(5 * time.Minute)},
	}
)

func SetCalculateExplorerStatistics(calculate bool) {
	calculateExplorerStatistics = calculate
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

func executeQuery(db *bun.DB, query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	_, err := db.Exec(query, ctx)
	return err
}

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}
