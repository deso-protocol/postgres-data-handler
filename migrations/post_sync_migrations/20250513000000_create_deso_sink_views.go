package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create a function that will be called when a transaction is committed.
		// For now, this is empty, but will still be called when a transaction is committed.
		// Other implementations of the postgres-data-handler may overwrite this function.
		// If this function has already been created, we don't want this placeholder to overwrite it.
		_, err := db.Exec(`
DO $$
BEGIN
	IF NOT EXISTS (select 1 from pg_views where schemaname='public' and viewname='token_balance_summary')
	THEN EXECUTE $view$
		create or replace view token_balance_summary
						(hodler_pkid, creator_pkid, unlocked_balance_nanos, locked_balance_nanos, total_balance) as
			with balance_entry_union as (select balance_entry.hodler_pkid  as hodler_pkid,
												balance_entry.creator_pkid as creator_pkid,
												balance_nanos              as unlocked_balance_nanos,
												0                          as locked_balance_nanos,
												balance_nanos              as total_balance_nanos,
												has_purchased              as has_purchased
										 from balance_entry
										 where is_dao_coin = true
										 union all
										 select locked_balance_entry.hodler_pkid  as hodler_pkid,
												locked_balance_entry.profile_pkid as creator_pkid,
												0                                 as unlocked_balance_nanos,
												balance_base_units                as locked_balance_nanos,
												balance_base_units                as total_balance_nanos,
												true                              as has_purchased
										 from locked_balance_entry)
			select hodler_pkid,
				   creator_pkid,
				   sum(unlocked_balance_nanos) AS unlocked_balance_nanos,
				   sum(locked_balance_nanos)   AS locked_balance_nanos,
				   sum(total_balance_nanos)    as total_balance,
				   bool_or(has_purchased)      as has_purchased
			from balance_entry_union
			group by hodler_pkid, creator_pkid
			$view$;
	END IF;
	IF NOT EXISTS (select 1 from pg_matviews where schemaname = 'public' and matviewname = 'token_balance_agg_v0')
	THEN EXECUTE $view$
		create materialized view token_balance_agg_v0 as
		SELECT balance_entry.creator_pkid,
			   sum(balance_entry.total_balance)          AS total_balance_nanos,
			   count(balance_entry.hodler_pkid)          AS hodler_count,
			   sum(balance_entry.locked_balance_nanos)   AS total_locked_balance_nanos,
			   sum(balance_entry.unlocked_balance_nanos) AS total_unlocked_balance_nanos
		FROM token_balance_summary balance_entry
		GROUP BY balance_entry.creator_pkid;

		create unique index on token_balance_agg_v0 (creator_pkid);
		COMMENT ON MATERIALIZED VIEW token_balance_agg_v0 IS E'@omit';
		$view$;
	END IF;
	IF NOT EXISTS (select 1 from pg_views where schemaname='public' and viewname='token_balance_agg')
	THEN EXECUTE $view$
		create or replace view token_balance_agg as
		select creator_pkid,
			   total_balance_nanos,
			   hodler_count,
			   total_locked_balance_nanos,
			   total_unlocked_balance_nanos
		from token_balance_agg_v0;
		$view$;
	END IF;
END $$;

CREATE TABLE deso_sinks_burn_txns
(
    transaction_hash varchar primary key,
    public_key       varchar not null,
    timestamp timestamp not null,
    index_in_block integer,
    block_height bigint
);

comment on table deso_sinks_burn_txns is E'@foreignKey (transaction_hash) references transaction (transaction_hash)|@foreignFieldName deso_sinks_burn_txn|@fieldName transaction'

CREATE TABLE deso_sinks_burn_amounts (
	public_key varchar primary key,
	total_coins_burned_nanos numeric default 0,
	last_update_block_height int
);

create or replace function refresh_deso_sinks_burn_amounts() returns void
                language plpgsql
            as
            $$
                DECLARE
                    latest_bh integer;
                BEGIN
            
                    --   Get the last refreshed block height. Be sure to skip unconfirmed reactions (too difficult to reconcile)
                    select least(
                                   (select coalesce(max(last_update_block_height), 0) from deso_sinks_burn_amounts),
                                   (select max(height) from block)
                               )
                      INTO latest_bh;
            
                    -- Upsert any new deso sinks. On conflict, add the new amounts to the existing amounts.
                    insert into deso_sinks_burn_amounts (public_key, total_coins_burned_nanos, last_update_block_height)
                    select
						public_key,
						sum(HEX_TO_NUMERIC(txn_meta ->> 'CoinsToBurnNanos')) as total_coins_burned_nanos,
						max(block_height) as last_update_block_height
					from transaction_partition_24
					where block_height > latest_bh AND ((
					public_key = 'BC1YLj3zNA7hRAqBVkvsTeqw7oi4H6ogKiAFL1VXhZy6pYeZcZ6TDRY'
					and (txn_meta ->> 'OperationType')::int = 1
					and (txn_meta ->> 'ProfilePublicKey')::varchar = 'A71P1dbtnB2hMpyHXh26kACeXEqq/k+AtM75Lfhwmls8')
					OR (
					public_key = 'BC1YLjEayZDjAPitJJX4Boy7LsEfN3sWAkYb3hgE9kGBirztsc2re1N'
					and (txn_meta ->> 'OperationType')::int = 1
					and (txn_meta ->> 'ProfilePublicKey')::varchar = 'A9VffrvFGTc7UV+8gyedQZO4WuTavYNOrDV1IElQ18x5'))
					GROUP BY public_key
                    on conflict (public_key) do update
            --             On conflict, add new burn amounts to existing burn amounts
                    set total_coins_burned_nanos = deso_sinks_burn_amounts.total_coins_burned_nanos + excluded.total_coins_burned_nanos,
                        last_update_block_height = excluded.last_update_block_height;
                	
					insert into deso_sinks_burn_txns (transaction_hash, public_key, timestamp, index_in_block, block_height)
					select transaction_hash,
						   public_key,
						   timestamp,
						   index_in_block,
						   block_height
					from transaction_partition_24
					where block_height > latest_bh AND ((
					public_key = 'BC1YLj3zNA7hRAqBVkvsTeqw7oi4H6ogKiAFL1VXhZy6pYeZcZ6TDRY'
					and (txn_meta ->> 'OperationType')::int = 1
					and (txn_meta ->> 'ProfilePublicKey')::varchar = 'A71P1dbtnB2hMpyHXh26kACeXEqq/k+AtM75Lfhwmls8')
					OR (
					public_key = 'BC1YLjEayZDjAPitJJX4Boy7LsEfN3sWAkYb3hgE9kGBirztsc2re1N'
					and (txn_meta ->> 'OperationType')::int = 1
					and (txn_meta ->> 'ProfilePublicKey')::varchar = 'A9VffrvFGTc7UV+8gyedQZO4WuTavYNOrDV1IElQ18x5'))
					ON CONFLICT (transaction_hash) DO NOTHING;
		
				END;
            $$;

-- Note that openfund currently does not return any public keys for the AMM. This view needs
-- to be updated anytime the amm public keys change, specifically the with amm_balances part.
create or replace view deso_sinks as
with amm_balances as (select *
                      from token_balance_summary tbs
                      where creator_pkid = 'BC1YLjEayZDjAPitJJX4Boy7LsEfN3sWAkYb3hgE9kGBirztsc2re1N'
                        and hodler_pkid = 'BC1YLghWqcM8kegcKTf7QGkT7Q4LLyMMZ2uvb7eY68HEiZjyj5v8Eh6')
select a.username,
       a.public_key,
       coalesce(tba.total_balance_nanos, 0)             as total_supply_nanos,
       coalesce(tba.hodler_count, 0)                    as holder_count,
       coalesce(tba.total_unlocked_balance_nanos, 0)    as unlocked_supply_nanos,
       coalesce(tba.total_locked_balance_nanos, 0)      as locked_supply_nanos,
       coalesce(dsba.total_coins_burned_nanos, 0)       as total_burn_nanos,
       coalesce(amm_balances.unlocked_balance_nanos, 0) as unlocked_amm_balance_nanos
from account a
         left join token_balance_agg tba
                   on a.pkid = tba.creator_pkid
         left join deso_sinks_burn_amounts dsba on a.public_key = dsba.public_key
         left join amm_balances on amm_balances.creator_pkid = a.pkid
where a.public_key in ('BC1YLj3zNA7hRAqBVkvsTeqw7oi4H6ogKiAFL1VXhZy6pYeZcZ6TDRY',
                       'BC1YLjEayZDjAPitJJX4Boy7LsEfN3sWAkYb3hgE9kGBirztsc2re1N');
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
