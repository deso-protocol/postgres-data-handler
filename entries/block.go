package entries

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type BlockEntry struct {
	BlockHash     string `pg:",pk,use_zero"`
	PrevBlockHash string
	TxnMerkleRoot string
	Timestamp     time.Time
	Height        uint64
	Nonce         uint64
	ExtraNonce    uint64
	BadgerKey     []byte `pg:",use_zero"`
}

type PGBlockEntry struct {
	bun.BaseModel `bun:"table:block"`
	BlockEntry
}

type TransactionEntry struct {
	TransactionHash              string `pg:",pk,use_zero"`
	BlockHash                    string
	Version                      uint16
	Inputs                       []map[string]string `bun:"type:jsonb"`
	Outputs                      []map[string]string `bun:"type:jsonb"`
	FeeNanos                     uint64
	NonceExperiationBlockHeight  uint64
	NoncePartialId               uint64
	TxnMeta                      lib.DeSoTxnMetadata `bun:"type:jsonb"`
	TxIndexMetadata              lib.DeSoEncoder     `bun:"type:jsonb"`
	TxIndexBasicTransferMetadata lib.DeSoEncoder     `bun:"type:jsonb"`
	TxnMetaBytes                 []byte
	TxnBytes                     []byte
	TxnType                      uint16
	PublicKey                    string
	ExtraData                    map[string]string `bun:"type:jsonb"`
	Signature                    []byte
	IndexInBlock                 uint64
}

type PGTransactionEntry struct {
	bun.BaseModel `bun:"table:transaction"`
	TransactionEntry
}

// Convert the UserAssociation DeSo encoder to the PG struct used by bun.
func BlockEncoderToPGStruct(block *lib.MsgDeSoBlock, keyBytes []byte) *PGBlockEntry {
	blockHash, _ := block.Hash()
	return &PGBlockEntry{
		BlockEntry: BlockEntry{
			BlockHash:     hex.EncodeToString(blockHash[:]),
			PrevBlockHash: hex.EncodeToString(block.Header.PrevBlockHash[:]),
			TxnMerkleRoot: hex.EncodeToString(block.Header.TransactionMerkleRoot[:]),
			Timestamp:     consumer.UnixNanoToTime(block.Header.TstampSecs * 1e9),
			Height:        block.Header.Height,
			Nonce:         block.Header.Nonce,
			ExtraNonce:    block.Header.ExtraNonce,
			BadgerKey:     keyBytes,
		},
	}
}

func TransactionEncoderToPGStruct(transaction *lib.MsgDeSoTxn, blockIndex uint64, blockHash string) (*PGTransactionEntry, error) {

	var txInputs []map[string]string
	for _, input := range transaction.TxInputs {
		txInputs = append(txInputs, map[string]string{
			"txid":  hex.EncodeToString(input.TxID[:]),
			"index": fmt.Sprintf("%d", input.Index),
		})
	}
	var txOutputs []map[string]string
	for _, output := range transaction.TxOutputs {
		txOutputs = append(txOutputs, map[string]string{
			"public_key":   hex.EncodeToString(output.PublicKey[:]),
			"amount_nanos": fmt.Sprintf("%d", output.AmountNanos),
		})
	}

	txnMetaBytes, err := transaction.TxnMeta.ToBytes(true)
	if err != nil {
		return nil, errors.Wrapf(err, "TransactionEncoderToPGStruct: Problem converting txn meta to bytes")
	}

	txnBytes, err := transaction.ToBytes(true)
	if err != nil {
		return nil, errors.Wrapf(err, "TransactionEncoderToPGStruct: Problem converting txn to bytes")
	}

	transactionEntry := &PGTransactionEntry{
		TransactionEntry: TransactionEntry{
			TransactionHash: hex.EncodeToString(transaction.Hash()[:]),
			BlockHash:       blockHash,
			Version:         uint16(transaction.TxnVersion),
			Inputs:          txInputs,
			Outputs:         txOutputs,
			FeeNanos:        transaction.TxnFeeNanos,
			TxnMeta:         transaction.TxnMeta,
			TxnMetaBytes:    txnMetaBytes,
			TxnBytes:        txnBytes,
			TxnType:         uint16(transaction.TxnMeta.GetTxnType()),
			PublicKey:       consumer.PublicKeyBytesToBase58Check(transaction.PublicKey[:]),
			ExtraData:       consumer.ExtraDataBytesToString(transaction.ExtraData),
			IndexInBlock:    blockIndex,
		},
	}

	if transaction.TxnNonce != nil {
		transactionEntry.NonceExperiationBlockHeight = transaction.TxnNonce.ExpirationBlockHeight
		transactionEntry.NoncePartialId = transaction.TxnNonce.PartialID
	}

	if transaction.Signature.Sign != nil {
		transactionEntry.Signature = transaction.Signature.ToBytes()
	}
	return transactionEntry, nil
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func BlockBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteBlockEntry(entries, db, operationType)
	} else {
		err = bulkInsertBlockEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertUtxoOperationsEntry inserts a batch of user_association entries into the database.
func bulkInsertBlockEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueBlocks := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgBlockEntrySlice := make([]*PGBlockEntry, 0)
	pgTransactionEntrySlice := make([]*PGTransactionEntry, 0)

	for _, entry := range uniqueBlocks {
		block := entry.Encoder.(*lib.MsgDeSoBlock)
		blockEntry := BlockEncoderToPGStruct(block, entry.KeyBytes)
		pgBlockEntrySlice = append(pgBlockEntrySlice, blockEntry)
		for jj, transaction := range block.Txns {
			pgTransactionEntry, err := TransactionEncoderToPGStruct(transaction, uint64(jj), blockEntry.BlockHash)
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

	transactionQuery := db.NewInsert().Model(&pgTransactionEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		transactionQuery = transactionQuery.On("CONFLICT (transaction_hash) DO UPDATE")
	}

	if _, err := transactionQuery.Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertTransaction: Error inserting entries")
	}
	return nil
}

// bulkDeleteBlockEntry deletes a batch of utxo_operation entries from the database.
func bulkDeleteBlockEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGBlockEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting entries")
	}

	return nil
}
