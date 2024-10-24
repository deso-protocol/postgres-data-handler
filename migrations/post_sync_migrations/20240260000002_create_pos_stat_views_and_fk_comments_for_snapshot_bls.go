package post_sync_migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`

CREATE OR REPLACE VIEW epoch_details_for_block as
select block_hash, epoch_number, bls.pkid as proposer_pkid
from block
         left join epoch_entry
                   on epoch_entry.initial_block_height <= block.height and
                      epoch_entry.final_block_height >= block.height
         left join bls_public_key_pkid_pair_snapshot_entry bls
                   on bls.snapshot_at_epoch_number = epoch_entry.snapshot_at_epoch_number and
                      block.proposer_voting_public_key = bls.bls_public_key;

		comment on view epoch_details_for_block is E'@unique block_hash\n@unique epoch_number\n@foreignKey (block_hash) references block (block_hash)|@foreignFieldName epochDetailForBlock|@fieldName block\n@foreignKey (epoch_number) references epoch_entry (epoch_number)|@foreignFieldName blockHashesInEpoch|@fieldName epochEntry\n@foreignKey (proposer_pkid) references account (pkid)|@foreignFieldName proposedBlockHashes|@fieldName proposer';
		comment on table bls_public_key_pkid_pair_snapshot_entry is E'@foreignKey (pkid) references account (pkid)|@foreignFieldName blsPublicKeyPkidPairSnapshotEntries|@fieldName account\n@foreignKey (snapshot_at_epoch_number) references epoch_entry (snapshot_at_epoch_number)|@foreignFieldName blsPublicKeyPkidPairSnapshotEntries|@fieldName epochEntry';
		comment on column bls_public_key_pkid_pair_snapshot_entry.badger_key is E'@omit';
`)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`	
			comment on column bls_public_key_pkid_pair_snapshot_entry.badger_key is NULL;	
			comment on table bls_public_key_pkid_pair_snapshot_entry is NULL; 
			DROP VIEW IF EXISTS epoch_details_for_block CASCADE;
`)
		if err != nil {
			return err
		}

		return nil
	})
}
