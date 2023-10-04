package post_sync_migrations

import (
	"context"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			CREATE INDEX transaction_hash_idx ON transaction_partitioned (transaction_hash);
			CREATE INDEX transaction_id_idx ON transaction_partitioned (transaction_id);
			CREATE INDEX transaction_block_height_idx ON transaction_partitioned (block_height desc);
			CREATE INDEX transaction_timestamp_idx ON transaction_partitioned (timestamp desc);
			CREATE INDEX transaction_index_in_block_idx ON transaction_partitioned (index_in_block);
			CREATE INDEX transaction_block_height_idx_in_block_idx ON transaction_partitioned (block_height desc, index_in_block asc);
			CREATE INDEX transaction_timestamp_idx_in_block_idx ON transaction_partitioned (timestamp desc, index_in_block asc);
			CREATE INDEX transaction_timestamp_idx_in_block_desc_idx ON transaction_partitioned (timestamp desc, index_in_block desc);
			CREATE INDEX transaction_index_badger_key_idx ON transaction_partitioned (badger_key);
			CREATE INDEX transaction_block_hash_index_idx ON transaction_partitioned (block_hash, index_in_block);
			CREATE INDEX transaction_block_hash_idx ON transaction_partitioned (block_hash);
			CREATE INDEX transaction_type_idx ON transaction_partitioned (txn_type);
			CREATE INDEX transaction_public_key_idx ON transaction_partitioned (public_key);
			CREATE INDEX transaction_public_key_timestamp_idx ON transaction_partitioned (public_key, timestamp desc);
			CREATE INDEX transaction_public_key_timestamp_index_in_block_idx ON transaction_partitioned (public_key, timestamp desc, index_in_block asc);
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE INDEX transaction_partition_27_tx_meta_assoc_type_idx ON transaction_partition_27 ((tx_index_metadata ->> 'AssociationType'));
			CREATE INDEX transaction_partition_27_tx_meta_assoc_value_idx ON transaction_partition_27 ((tx_index_metadata ->> 'AssociationValue'));
			CREATE INDEX transaction_partition_27_tx_meta_app_pub_key_idx ON transaction_partition_27 ((tx_index_metadata ->> 'AppPublicKeyBase58Check'));
			CREATE INDEX transaction_partition_27_tx_meta_target_pub_key_idx ON transaction_partition_27 ((tx_index_metadata ->> 'TargetUserPublicKeyBase58Check'));
			
			CREATE INDEX transaction_partition_26_tx_meta_buying_dao_coin_pub_key_idx ON transaction_partition_29 ((tx_index_metadata ->> 'BuyingDAOCoinCreatorPublicKey'));
			CREATE INDEX transaction_partition_26_tx_meta_selling_dao_coin_pub_key_idx ON transaction_partition_29 ((tx_index_metadata ->> 'SellingDAOCoinCreatorPublicKey'));


			CREATE INDEX transaction_partition_29_tx_meta_assoc_type_idx ON transaction_partition_29 ((tx_index_metadata ->> 'AssociationType'));
			CREATE INDEX transaction_partition_29_tx_meta_assoc_value_idx ON transaction_partition_29 ((tx_index_metadata ->> 'AssociationValue'));
			CREATE INDEX transaction_partition_29_tx_meta_app_pub_key_idx ON transaction_partition_29 ((tx_index_metadata ->> 'AppPublicKeyBase58Check'));
			CREATE INDEX transaction_partition_29_tx_meta_target_pub_key_idx ON transaction_partition_29 ((tx_index_metadata ->> 'PostHashHex'));
			
			CREATE INDEX transaction_partition_17_tx_meta_nft_post_hash_idx ON transaction_partition_17 ((tx_index_metadata ->> 'NFTPostHashHex'));
			CREATE INDEX transaction_partition_17_tx_meta_nft_sn_idx ON transaction_partition_17 ((tx_index_metadata ->> 'SerialNumber'));
			
			CREATE INDEX transaction_partition_10_tx_meta_is_unlike_idx ON transaction_partition_10 ((tx_index_metadata ->> 'IsUnlike'));
			CREATE INDEX transaction_partition_10_tx_meta_is_unlike_false_idx
			ON transaction_partition_10 ((tx_index_metadata ->> 'IsUnlike'))
			WHERE (tx_index_metadata ->> 'IsUnlike') = 'false';
			CREATE INDEX transaction_partition_10_tx_meta_post_hash_idx ON transaction_partition_10 ((tx_index_metadata ->> 'PostHashHex'));

			CREATE EXTENSION IF NOT EXISTS pgcrypto;
			
			CREATE OR REPLACE FUNCTION checksum(bytea) RETURNS bytea AS $$
			BEGIN
				RETURN substring(digest(digest($1, 'sha256'), 'sha256') from 1 for 4);
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;
			
			CREATE OR REPLACE FUNCTION base58_check_encode_with_prefix(input bytea) RETURNS TEXT AS $$
			DECLARE
				prefix bytea := E'\\xcd1400'::bytea;
				b bytea;
				big_val NUMERIC;
			BEGIN
				b := prefix || input || checksum(prefix || input);
			
				-- Convert bytea to a big numeric for Base58 encoding
				SELECT INTO big_val bytes_to_bigint(b);
				RETURN base58_encode(big_val);
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;
			
			CREATE OR REPLACE FUNCTION bytes_to_bigint(byte_data bytea) RETURNS numeric AS $$
			DECLARE
				len int;
				i int;
				val numeric := 0;
				byte_val smallint;
			BEGIN
				len := LENGTH(byte_data);
			
				FOR i IN 1..len LOOP
					byte_val := GET_BYTE(byte_data, i-1);
					val := val * 256 + byte_val;
				END LOOP;
			
				RETURN val;
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;

			create or replace function int_to_bytea(i integer) returns bytea
				language plpgsql
			as
			$$
			BEGIN
				RETURN decode(lpad(to_hex(i),2,'0'), 'hex');
			END;
			$$;
			    
			create or replace function jsonb_to_bytea(j jsonb) returns bytea
				language plpgsql
			as
			$$
			DECLARE
				res bytea := E'';
				val text;
			BEGIN
				FOR val IN SELECT jsonb_array_elements_text(j)
				LOOP
					res := res || int_to_bytea(val::int);
				END LOOP;
				RETURN res;
			END;
			$$;
			
			CREATE OR REPLACE FUNCTION base58_encode(num NUMERIC)
			  RETURNS VARCHAR(255) AS $encoded$
			
			DECLARE
			  alphabet   VARCHAR(255);
			  base_count NUMERIC;
			  encoded    VARCHAR(255);
			  divisor    NUMERIC;
			  mod        NUMERIC;
			
			BEGIN
			  alphabet := '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
			  base_count := 58;
			  encoded := '';
			
			  WHILE num > 0 LOOP
				divisor := floor(num / base_count);
				mod := num - (divisor * base_count);
			
				-- Adjust for negative remainder
				IF mod < 0 THEN
				  divisor := divisor - 1;
				  mod := mod + base_count;
				END IF;
			
				encoded := concat(substring(alphabet FROM mod::INT + 1 FOR 1), encoded);
				num := divisor;
			  END LOOP;
			
			  RETURN encoded;
			END; $encoded$
			LANGUAGE PLPGSQL IMMUTABLE;
			
			CREATE OR REPLACE FUNCTION base64_to_base58(base58_string varchar) RETURNS TEXT AS $$
			BEGIN
				return base58_check_encode_with_prefix(decode(base58_string, 'base64'));
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;
			    
			CREATE INDEX transaction_partition_11_tx_meta_prof_pub_key_idx ON transaction_partition_11 (base64_to_base58(txn_meta ->> 'ProfilePublicKey'));
			
			CREATE INDEX transaction_partition_02_tx_meta_post_hash_idx on transaction_partition_02 ((tx_index_metadata ->> 'PostHashHex'));
			CREATE INDEX transaction_partition_02_tx_meta_diamond_level_idx on transaction_partition_02 ((tx_index_metadata ->> 'DiamondLevel'));

			CREATE EXTENSION if not exists btree_gin;
			
			CREATE INDEX idx_transaction_creator_key
			ON transaction_partition_17
			USING gin ((tx_index_metadata->'NFTRoyaltiesMetadata'->>'CreatorPublicKeyBase58Check'));
			
			CREATE INDEX idx_transaction_additional_royalties
			ON transaction_partition_17
			USING gin ((tx_index_metadata->'NFTRoyaltiesMetadata'->'AdditionalDESORoyaltiesMap'));

			CREATE INDEX transaction_partition_18_tx_meta_is_buy_now_idx on transaction_partition_18 ((tx_index_metadata ->> 'IsBuyNowBid' = 'true'));

			CREATE INDEX transaction_partition_26_buying_coin_pub_key_idx ON transaction_partition_26 ((tx_index_metadata ->> 'BuyingDAOCoinCreatorPublicKey'));
			CREATE INDEX transaction_partition_26_selling_coin_pub_key_idx ON transaction_partition_26 ((tx_index_metadata ->> 'SellingDAOCoinCreatorPublicKey'));
			CREATE INDEX transaction_partition_26_gin_tx_index_metadata_idx ON transaction_partition_26 USING gin (tx_index_metadata jsonb_path_ops);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP function IF EXISTS base64_to_base58;
			DROP function IF EXISTS base58_check_encode_with_prefix;
			DROP function IF EXISTS base58_encode;
			DROP function IF EXISTS bytes_to_bigint;
			DROP function IF EXISTS checksum;
			DROP function IF EXISTS jsonb_to_bytea;
			DROP EXTENSION IF EXISTS pgcrypto;
			DROP EXTENSION IF EXISTS btree_gin;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
