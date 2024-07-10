package entries

import (
	"context"
	"encoding/hex"
	"reflect"
	"time"

	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type BlockEntry struct {
	BlockHash                    string `pg:",pk,use_zero"`
	PrevBlockHash                string
	TxnMerkleRoot                string
	Timestamp                    time.Time
	Height                       uint64
	Nonce                        uint64
	ExtraNonce                   uint64
	BlockVersion                 uint32
	ProposerVotingPublicKey      string `pg:",use_zero"`
	ProposerRandomSeedSignature  string `pg:",use_zero"`
	ProposedInView               uint64
	ProposerVotePartialSignature string `pg:",use_zero"`
	// TODO: Quorum Certificates. Separate entry.

	BadgerKey []byte `pg:",use_zero"`
}

type PGBlockEntry struct {
	bun.BaseModel `bun:"table:block"`
	BlockEntry
}

type BlockSigner struct {
	BlockHash   string
	SignerIndex uint64
}

type PGBlockSigner struct {
	bun.BaseModel `bun:"table:block_signer"`
	BlockSigner
}

// Convert the UserAssociation DeSo encoder to the PG struct used by bun.
func BlockEncoderToPGStruct(block *lib.MsgDeSoBlock, keyBytes []byte, params *lib.DeSoParams) (*PGBlockEntry, []*PGBlockSigner) {
	blockHash, _ := block.Hash()
	blockHashHex := hex.EncodeToString(blockHash[:])
	qc := block.Header.GetQC()
	blockSigners := []*PGBlockSigner{}
	if !isInterfaceNil(qc) {
		aggSig := qc.GetAggregatedSignature()
		if !isInterfaceNil(aggSig) {
			signersList := aggSig.GetSignersList()
			for ii := 0; ii < signersList.Size(); ii++ {
				// Skip signers that didn't sign.
				if !signersList.Get(ii) {
					continue
				}
				blockSigners = append(blockSigners, &PGBlockSigner{
					BlockSigner: BlockSigner{
						BlockHash:   blockHashHex,
						SignerIndex: uint64(ii),
					},
				})
			}
		}
	}
	return &PGBlockEntry{
		BlockEntry: BlockEntry{
			BlockHash:                    blockHashHex,
			PrevBlockHash:                hex.EncodeToString(block.Header.PrevBlockHash[:]),
			TxnMerkleRoot:                hex.EncodeToString(block.Header.TransactionMerkleRoot[:]),
			Timestamp:                    consumer.UnixNanoToTime(uint64(block.Header.TstampNanoSecs)),
			Height:                       block.Header.Height,
			Nonce:                        block.Header.Nonce,
			ExtraNonce:                   block.Header.ExtraNonce,
			BlockVersion:                 block.Header.Version,
			ProposerVotingPublicKey:      block.Header.ProposerVotingPublicKey.ToString(),
			ProposerRandomSeedSignature:  block.Header.ProposerRandomSeedSignature.ToString(),
			ProposedInView:               block.Header.ProposedInView,
			ProposerVotePartialSignature: block.Header.ProposerVotePartialSignature.ToString(),
			BadgerKey:                    keyBytes,
		},
	}, blockSigners
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func BlockBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteBlockEntry(entries, db, operationType)
	} else {
		err = bulkInsertBlockEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertUtxoOperationsEntry inserts a batch of user_association entries into the database.
func bulkInsertBlockEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// If this block is a part of the initial sync, skip it - it will be handled by the utxo operations.
	if operationType == lib.DbOperationTypeInsert {
		return nil
	}

	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueBlocks := consumer.UniqueEntries(entries)

	// Create a new array to hold the bun struct.
	pgBlockEntrySlice := make([]*PGBlockEntry, 0)
	pgTransactionEntrySlice := make([]*PGTransactionEntry, 0)
	pgBlockSignersEntrySlice := make([]*PGBlockSigner, 0)

	for _, entry := range uniqueBlocks {
		block := entry.Encoder.(*lib.MsgDeSoBlock)
		blockEntry, blockSigners := BlockEncoderToPGStruct(block, entry.KeyBytes, params)
		pgBlockEntrySlice = append(pgBlockEntrySlice, blockEntry)
		pgBlockSignersEntrySlice = append(pgBlockSignersEntrySlice, blockSigners...)
		for jj, transaction := range block.Txns {
			indexInBlock := uint64(jj)
			pgTransactionEntry, err := TransactionEncoderToPGStruct(
				transaction,
				&indexInBlock,
				blockEntry.BlockHash,
				blockEntry.Height,
				blockEntry.Timestamp,
				nil,
				nil,
				params,
			)
			if err != nil {
				return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Problem converting transaction to PG struct")
			}
			pgTransactionEntrySlice = append(pgTransactionEntrySlice, pgTransactionEntry)
			if transaction.TxnMeta.GetTxnType() != lib.TxnTypeAtomicTxnsWrapper {
				continue
			}
			innerTxns, err := parseInnerTxnsFromAtomicTxn(pgTransactionEntry, params)
			if err != nil {
				return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Problem parsing inner txns from atomic txn")
			}
			pgTransactionEntrySlice = append(pgTransactionEntrySlice, innerTxns...)
		}
	}

	blockQuery := db.NewInsert().Model(&pgBlockEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		blockQuery = blockQuery.On(`CONFLICT (height) DO UPDATE
			SET prev_block_hash = EXCLUDED.prev_block_hash,
				block_hash = EXCLUDED.block_hash,
				txn_merkle_root = EXCLUDED.txn_merkle_root,
				timestamp = EXCLUDED.timestamp,
				nonce = EXCLUDED.nonce,
				extra_nonce = EXCLUDED.extra_nonce,
				block_version = EXCLUDED.block_version,
				proposer_voting_public_key = EXCLUDED.proposer_voting_public_key,
				proposer_random_seed_signature = EXCLUDED.proposer_random_seed_signature,
				proposed_in_view = EXCLUDED.proposed_in_view,
				proposer_vote_partial_signature = EXCLUDED.proposer_vote_partial_signature
				WHERE handle_block_conflict(pg_block_entry.block_hash) IS NOT NULL
		`)
	}

	if _, err := blockQuery.Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBlock: Error inserting entries")
	}

	if err := bulkInsertTransactionEntry(pgTransactionEntrySlice, db, operationType); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBlock: Error inserting transaction entries")
	}

	if len(pgBlockSignersEntrySlice) > 0 {
		// Execute the insert query.
		query := db.NewInsert().Model(&pgBlockSignersEntrySlice)

		if operationType == lib.DbOperationTypeUpsert {
			query = query.On("CONFLICT (block_hash, signer_index) DO UPDATE")
		}

		if _, err := query.Returning("").Exec(context.Background()); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Error inserting block signers")
		}
	}

	return nil
}

// bulkDeleteBlockEntry deletes a batch of block entries from the database.
func bulkDeleteBlockEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	return bulkDeleteBlockEntriesFromKeysToDelete(db, keysToDelete)
}

// bulkDeleteBlockEntriesFromKeysToDelete deletes a batch of block entries from the database.
// It also deletes any transactions and utxo operations associated with the block.
func bulkDeleteBlockEntriesFromKeysToDelete(db bun.IDB, keysToDelete [][]byte) error {
	// Execute the delete query on the blocks table.
	if _, err := db.NewDelete().
		Model(&PGBlockEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting entries")
	}
	// Get block hashes from keys to delete.
	blockHashHexesToDelete := make([]string, len(keysToDelete))
	for ii, keyToDelete := range keysToDelete {
		blockHashHexesToDelete[ii] = hex.EncodeToString(consumer.GetBlockHashBytesFromKey(keyToDelete))
	}

	// Delete any transactions associated with the block.
	if _, err := db.NewDelete().
		Model(&PGTransactionEntry{}).
		Where("block_hash IN (?)", bun.In(blockHashHexesToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting transaction entries")
	}

	// Delete any utxo operations associated with the block.
	if _, err := db.NewDelete().
		Model(&PGUtxoOperationEntry{}).
		Where("block_hash IN (?)", bun.In(blockHashHexesToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting utxo operation entries")
	}

	// Delete any signers associated with the block.
	if _, err := db.NewDelete().
		Model(&PGBlockSigner{}).
		Where("block_hash IN (?)", bun.In(blockHashHexesToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting block signers")
	}
	return nil
}

// golang interface types are stored as a tuple of (type, value). A single i==nil check is not enough to
// determine if a pointer that implements an interface is nil. This function checks if the interface is nil
// by checking if the pointer itself is nil.
func isInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}

	value := reflect.ValueOf(i)
	return value.Kind() == reflect.Ptr && value.IsNil()
}
