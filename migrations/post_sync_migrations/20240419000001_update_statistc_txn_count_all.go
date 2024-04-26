package post_sync_migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {

		err := RunMigrationWithRetries(db, fmt.Sprintf(`
			DROP MATERIALIZED VIEW IF EXISTS statistic_txn_count_all CASCADE;
			CREATE MATERIALIZED VIEW statistic_txn_count_all AS
			SELECT SUM(get_transaction_count(s.i)) as count,
			       0 as id
			FROM generate_series(1, 44) AS s(i);

            CREATE UNIQUE INDEX statistic_txn_count_all_unique_index ON statistic_txn_count_all (id);
			comment on materialized view statistic_txn_count_all is E'@omit';
			%v
`, buildStatisticsView()))
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		err := RunMigrationWithRetries(db, fmt.Sprintf(`
			DROP MATERIALIZED VIEW IF EXISTS statistic_txn_count_all CASCADE;		
			CREATE MATERIALIZED VIEW statistic_txn_count_all AS
			SELECT SUM(get_transaction_count(s.i)) as count,
			       0 as id
			FROM generate_series(1, 33) AS s(i);

            CREATE UNIQUE INDEX statistic_txn_count_all_unique_index ON statistic_txn_count_all (id);
			comment on materialized view statistic_txn_count_all is E'@omit';
			%v
`, buildStatisticsView()))
		if err != nil {
			return err
		}
		return nil
	})
}
func buildStatisticsView() string {
	return `
CREATE VIEW statistic_dashboard AS
			SELECT
				statistic_txn_count_all.count as txn_count_all,
				statistic_txn_count_30_d.count as txn_count_30_d,
				statistic_wallet_count_all.count as wallet_count_all,
				statistic_active_wallet_count_30_d.count as active_wallet_count_30_d,
				statistic_new_wallet_count_30_d.count as new_wallet_count_30_d,
				statistic_block_height_current.height as block_height_current,
				statistic_txn_count_pending.count as txn_count_pending,
				statistic_txn_fee_1_d.avg as txn_fee_1_d,
				statistic_total_supply.sum as total_supply,
				statistic_post_count.count as post_count,
				statistic_post_longform_count.count as post_longform_count,
				statistic_comment_count.count as comment_count,
				statistic_repost_count.count as repost_count,
				statistic_txn_count_creator_coin.count as txn_count_creator_coin,
				statistic_txn_count_nft.count as txn_count_nft,
				statistic_txn_count_dex.count as txn_count_dex,
				statistic_txn_count_social.count as txn_count_social,
				statistic_follow_count.count as follow_count,
				statistic_message_count.count as message_count
			FROM
			statistic_txn_count_all
			CROSS JOIN
			statistic_txn_count_30_d
			CROSS JOIN
			statistic_wallet_count_all
			CROSS JOIN
			statistic_active_wallet_count_30_d
			CROSS JOIN
			statistic_new_wallet_count_30_d
			CROSS JOIN
			statistic_block_height_current
			CROSS JOIN
			statistic_txn_count_pending
			CROSS JOIN
			statistic_txn_fee_1_d
			CROSS JOIN
			statistic_total_supply
			CROSS JOIN
			statistic_post_count
			CROSS JOIN
			statistic_post_longform_count
			CROSS JOIN
			statistic_comment_count
			CROSS JOIN
			statistic_repost_count
			CROSS JOIN
			statistic_txn_count_creator_coin
			CROSS JOIN
			statistic_txn_count_nft
			CROSS JOIN
			statistic_txn_count_dex
			CROSS JOIN
			statistic_txn_count_social
			CROSS JOIN
			statistic_follow_count
			CROSS JOIN
			statistic_message_count;
			comment on view statistic_dashboard is E'@name dashboardStat';
`
}
