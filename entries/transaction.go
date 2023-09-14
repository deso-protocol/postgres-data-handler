package entries

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/deso-protocol/backend/routes"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type TransactionEntry struct {
	TransactionHash              string `pg:",pk,use_zero"`
	TransactionId                string `pg:",use_zero"`
	BlockHash                    string
	Version                      uint16
	Inputs                       []map[string]string `bun:"type:jsonb"`
	Outputs                      []map[string]string `bun:"type:jsonb"`
	FeeNanos                     uint64
	NonceExpirationBlockHeight   uint64
	NoncePartialId               uint64
	TxnMeta                      lib.DeSoTxnMetadata        `bun:"type:jsonb"`
	TxnMetaResponse              routes.TransactionResponse `bun:"type:jsonb"`
	TxIndexMetadata              lib.DeSoEncoder            `bun:"type:jsonb"`
	TxIndexBasicTransferMetadata lib.DeSoEncoder            `bun:"type:jsonb"`
	TxnMetaBytes                 []byte
	TxnBytes                     []byte
	TxnType                      uint16
	PublicKey                    string
	ExtraData                    map[string]string `bun:"type:jsonb"`
	Signature                    []byte
	IndexInBlock                 uint64
	BlockHeight                  uint64
	Timestamp                    time.Time `pg:",use_zero"`
	BadgerKey                    []byte    `pg:",use_zero"`
}

type PGTransactionEntry struct {
	bun.BaseModel `bun:"table:transaction_partitioned"`
	TransactionEntry
}

func TransactionEncoderToPGStruct(transaction *lib.MsgDeSoTxn, blockIndex uint64, blockHash string, blockHeight uint64, timestamp time.Time, params *lib.DeSoParams) (*PGTransactionEntry, error) {

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
			"public_key":   consumer.PublicKeyBytesToBase58Check(output.PublicKey[:], params),
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
			TransactionId:   consumer.PublicKeyBytesToBase58Check(transaction.Hash()[:], params),
			BlockHash:       blockHash,
			Version:         uint16(transaction.TxnVersion),
			Inputs:          txInputs,
			Outputs:         txOutputs,
			FeeNanos:        transaction.TxnFeeNanos,
			TxnMeta:         transaction.TxnMeta,
			TxnMetaBytes:    txnMetaBytes,
			TxnBytes:        txnBytes,
			TxnType:         uint16(transaction.TxnMeta.GetTxnType()),
			PublicKey:       consumer.PublicKeyBytesToBase58Check(transaction.PublicKey[:], params),
			ExtraData:       consumer.ExtraDataBytesToString(transaction.ExtraData),
			IndexInBlock:    blockIndex,
			BlockHeight:     blockHeight,
			Timestamp:       timestamp,
			BadgerKey:       transaction.Hash()[:],
		},
	}

	if transaction.TxnNonce != nil {
		transactionEntry.NonceExpirationBlockHeight = transaction.TxnNonce.ExpirationBlockHeight
		transactionEntry.NoncePartialId = transaction.TxnNonce.PartialID
	}

	if transaction.Signature.Sign != nil {
		transactionEntry.Signature = transaction.Signature.ToBytes()
	}
	return transactionEntry, nil
}

// TransactionBatchOperation is the entry point for processing a batch of transaction entries. It determines the appropriate handler
// based on the operation type and executes it.
func TransactionBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteTransactionEntry(entries, db, operationType)
	} else {
		err = transformAndBulkInsertTransactionEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

func transformTransactionEntry(entries []*lib.StateChangeEntry, params *lib.DeSoParams) ([]*PGTransactionEntry, error) {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueTransactions := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgTransactionEntrySlice := make([]*PGTransactionEntry, 0)

	for _, entry := range uniqueTransactions {
		transaction := entry.Encoder.(*lib.MsgDeSoTxn)
		transactionEntry, err := TransactionEncoderToPGStruct(transaction, 0, "", 0, time.Now(), params)
		if err != nil {
			return nil, errors.Wrapf(err, "entries.transformAndBulkInsertTransactionEntry: Problem converting transaction to PG struct")
		}
		pgTransactionEntrySlice = append(pgTransactionEntrySlice, transactionEntry)
	}
	return pgTransactionEntrySlice, nil
}

func bulkInsertTransactionEntry(entries []*PGTransactionEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Bulk insert the entries.
	transactionQuery := db.NewInsert().Model(&entries)

	if operationType == lib.DbOperationTypeUpsert {
		transactionQuery = transactionQuery.On("CONFLICT (transaction_hash, txn_type) DO UPDATE")
	}

	if _, err := transactionQuery.Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertTransaction: Error inserting entries")
	}
	return nil
}

// transformAndBulkInsertTransactionEntry inserts a batch of user_association entries into the database.
func transformAndBulkInsertTransactionEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	pgTransactionEntrySlice, err := transformTransactionEntry(entries, params)
	if err != nil {
		return errors.Wrapf(err, "entries.transformAndBulkInsertTransactionEntry: Problem transforming transaction entries")
	}

	err = bulkInsertTransactionEntry(pgTransactionEntrySlice, db, operationType)

	if err != nil {
		return errors.Wrapf(err, "entries.transformAndBulkInsertTransactionEntry: Problem inserting transaction entries")
	}
	return nil
}

// bulkDeleteTransactionEntry deletes a batch of transaction entries from the database.
func bulkDeleteTransactionEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGTransactionEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBlockEntry: Error deleting entries")
	}

	return nil
}
