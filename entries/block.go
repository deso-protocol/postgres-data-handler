package entries

import (
	"context"
	"encoding/hex"
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

// Convert the UserAssociation DeSo encoder to the PG struct used by bun.
func BlockEncoderToPGStruct(block *lib.MsgDeSoBlock, keyBytes []byte, params *lib.DeSoParams) *PGBlockEntry {
	blockHash, _ := block.Hash()
	return &PGBlockEntry{
		BlockEntry: BlockEntry{
			BlockHash:                    hex.EncodeToString(blockHash[:]),
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
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func BlockBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
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
func bulkInsertBlockEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// If this block is a part of the initial sync, skip it - it will be handled by the utxo operations.
	if operationType == lib.DbOperationTypeInsert {
		return nil
	}

	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueBlocks := consumer.UniqueEntries(entries)

	// We need to check if this block is replacing an existing block at the same height.
	// If it is, we need to delete the existing block and all transactions associated with it.
	// Get all block with matching heights and different hashes.
	heights := make([]uint64, len(entries))
	hashes := make([]string, len(entries))
	for ii, entry := range uniqueBlocks {
		heights[ii] = entry.Encoder.(*lib.MsgDeSoBlock).Header.Height
		hash, err := entry.Encoder.(*lib.MsgDeSoBlock).Hash()
		if err != nil {
			return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Error getting block hash")
		}
		hashes[ii] = hex.EncodeToString(hash[:])
	}
	blocks := []*PGBlockEntry{}
	err := db.NewSelect().
		Model(&blocks).
		Where("height IN (?)", bun.In(heights)).
		Where("block_hash NOT IN (?)", bun.In(hashes)).
		Scan(context.Background())
	if err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Error getting blocks")
	}
	// If we have blocks at the same height, delete them and their transactions.
	if len(blocks) > 0 {
		keysToDelete := make([][]byte, len(blocks))
		for ii, block := range blocks {
			keysToDelete[ii] = block.BadgerKey
		}
		if err = bulkDeleteBlockEntriesFromKeysToDelete(db, keysToDelete); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Error deleting blocks")
		}
	}

	// Create a new array to hold the bun struct.
	pgBlockEntrySlice := make([]*PGBlockEntry, 0)
	pgTransactionEntrySlice := make([]*PGTransactionEntry, 0)

	for _, entry := range uniqueBlocks {
		block := entry.Encoder.(*lib.MsgDeSoBlock)
		blockEntry := BlockEncoderToPGStruct(block, entry.KeyBytes, params)
		pgBlockEntrySlice = append(pgBlockEntrySlice, blockEntry)
		for jj, transaction := range block.Txns {
			pgTransactionEntry, err := TransactionEncoderToPGStruct(
				transaction, uint64(jj), blockEntry.BlockHash, blockEntry.Height, blockEntry.Timestamp, params,
			)
			if err != nil {
				return errors.Wrapf(err, "entries.bulkInsertBlockEntry: Problem converting transaction to PG struct")
			}
			pgTransactionEntrySlice = append(pgTransactionEntrySlice, pgTransactionEntry)
		}
	}

	blockQuery := db.NewInsert().Model(&pgBlockEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		blockQuery = blockQuery.On("CONFLICT (block_hash) DO UPDATE")
	}

	if _, err := blockQuery.Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBlock: Error inserting entries")
	}

	if err = bulkInsertTransactionEntry(pgTransactionEntrySlice, db, operationType); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBlock: Error inserting transaction entries")
	}

	return nil
}

// bulkDeleteBlockEntry deletes a batch of block entries from the database.
func bulkDeleteBlockEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	return bulkDeleteBlockEntriesFromKeysToDelete(db, keysToDelete)
}

// bulkDeleteBlockEntriesFromKeysToDelete deletes a batch of block entries from the database.
// It also deletes any transactions and utxo operations associated with the block.
func bulkDeleteBlockEntriesFromKeysToDelete(db *bun.DB, keysToDelete [][]byte) error {
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
	return nil
}
