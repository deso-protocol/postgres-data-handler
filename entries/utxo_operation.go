package entries

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"sync"
	"time"
)

type UtxoOperation struct {
	UtxoOpEntryType  string `pg:",use_zero"`
	UtxoOpIndex      uint64 `pg:",use_zero"`
	TransactionIndex uint64 `pg:",use_zero"`
	ArrayIndex       uint64 `pg:",use_zero"`
	BlockHash        string `pg:",use_zero"`
}

type UtxoOperationEntry struct {
	OperationType    uint
	TransactionIndex uint64
	UtxoOpIndex      uint64
	BlockHash        string
	UtxoOpBytes      []byte
}

type PGUtxoOperationEntry struct {
	bun.BaseModel `bun:"table:utxo_operation"`
	UtxoOperationEntry
}

type AffectedPublicKeyEntry struct {
	PublicKey       string    `pg:",pk,use_zero"`
	Metadata        string    `pg:",pk,use_zero"`
	Timestamp       time.Time `pg:",use_zero"`
	TransactionHash string    `pg:",pk,use_zero"`
}

type PGAffectedPublicKeyEntry struct {
	bun.BaseModel `bun:"table:affected_public_key"`
	AffectedPublicKeyEntry
}

// Convert the UserAssociation DeSo encoder to the PG struct used by bun.
func UtxoOperationEncoderToPGStruct(utxoOperationEntry *lib.UtxoOperation, keyBytes []byte, transactionIndex uint64, utxoOpIndex uint64, blockHeight uint64) *PGUtxoOperationEntry {
	return &PGUtxoOperationEntry{
		UtxoOperationEntry: UtxoOperationEntry{
			OperationType:    uint(utxoOperationEntry.Type),
			TransactionIndex: transactionIndex,
			UtxoOpIndex:      utxoOpIndex,
			BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(keyBytes),
			UtxoOpBytes:      lib.EncodeToBytes(blockHeight, utxoOperationEntry),
		},
	}
}

func ConvertUtxoOperationKeyToBlockHashHex(keyBytes []byte) string {
	return hex.EncodeToString(keyBytes[1:])
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func UtxoOperationBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteUtxoOperationEntry(entries, db, operationType)
	} else {
		err = bulkInsertUtxoOperationsEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertUtxoOperationsEntry inserts a batch of utxo operation entries into the database.
func bulkInsertUtxoOperationsEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {

	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transactions added to this slice will have their txindex metadata updated.
	transactionUpdates := make([]*PGTransactionEntry, 0)
	affectedPublicKeys := make([]*PGAffectedPublicKeyEntry, 0)

	// Loop through the utxo op bundles and extract the utxo operation entries from them.
	for _, entry := range uniqueEntries {

		// Check if the entry is a bundle with multiple utxo operations, or a single transaction.
		// If bundle, get a list of transactions based on the block hash extracted from the key.
		// If single transaction, get the transaction based on the transaction hash, extracted from the key.

		transactions := []*PGTransactionEntry{}
		// We can use this function regardless of the db prefix, because both block_hash and transaction_hash
		// are stored in the same blockHashHex format in the key.
		blockHash := ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes)
		filterField := ""

		// Determine how the transactions should be filtered based on the entry key prefix.
		if bytes.Equal(entry.KeyBytes[:1], lib.Prefixes.PrefixTxnHashToUtxoOps) {
			filterField = "transaction_hash"
		} else if bytes.Equal(entry.KeyBytes[:1], lib.Prefixes.PrefixBlockHashToUtxoOperations) {
			filterField = "block_hash"
		} else {
			return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Unrecognized prefix %v", entry.KeyBytes[:1])
		}

		// Note: it's normally considered bad practice to use string formatting to insert values into a query. However,
		// in this case, the filterField is a constant and the value is clearly only block hash or transaction hash -
		// so there is no risk of SQL injection.
		err := db.NewSelect().Model(&transactions).Column("txn_bytes", "transaction_hash").Where(fmt.Sprintf("%s = ?", filterField), blockHash).Order("index_in_block ASC").Scan(context.Background())
		if err != nil {
			return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem getting transactions for entry %+v at block height %v", entry, entry.BlockHeight)
		}

		utxoOperations, ok := entry.Encoder.(*lib.UtxoOperationBundle)
		if !ok {
			return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem with entry %v", entry)
		}

		// Limit the number of concurrent txindex threads to avoid overloading the CPU
		const maxConcurrency = 50
		maxConcurrencySemaphore := make(chan bool, maxConcurrency)

		// Create a wait group to wait for all the goroutines to finish.
		var wg sync.WaitGroup
		wg.Add(len(utxoOperations.UtxoOpBundle))

		for jj := range utxoOperations.UtxoOpBundle {
			maxConcurrencySemaphore <- true
			go func(idx int) {
				// Defer the wait group so we can track when this goroutine is done.
				defer wg.Done()
				defer func() { <-maxConcurrencySemaphore }() // Release from the semaphore when done.

				utxoOps := utxoOperations.UtxoOpBundle[idx]
				jj := idx
				// Update the transaction metadata for this transaction.
				if jj < len(transactions) {
					transaction := &lib.MsgDeSoTxn{}
					err = transaction.FromBytes(transactions[jj].TxnBytes)
					if err != nil {
						fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Problem decoding transaction for entry %+v at block height %v", entry, entry.BlockHeight)
						return
					}
					txIndexMetadata, err := consumer.ComputeTransactionMetadata(transaction, blockHash, &lib.DeSoMainnetParams, transaction.TxnFeeNanos, uint64(jj), utxoOps)
					if err != nil {
						fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Problem computing transaction metadata for entry %+v at block height %v", entry, entry.BlockHeight)
						return
					}

					metadata := txIndexMetadata.GetEncoderForTxType(transaction.TxnMeta.GetTxnType())
					basicTransferMetadata := txIndexMetadata.BasicTransferTxindexMetadata
					basicTransferMetadata.UtxoOps = nil

					transactions[jj].TxIndexMetadata = metadata

					transactions[jj].TxIndexBasicTransferMetadata = txIndexMetadata.GetEncoderForTxType(lib.TxnTypeBasicTransfer)

					affectedPublicKeySet := make(map[string]bool)
					for _, affectedPublicKey := range txIndexMetadata.AffectedPublicKeys {
						if _, ok := affectedPublicKeySet[affectedPublicKey.PublicKeyBase58Check]; ok {
							continue
						}
						affectedPublicKeySet[affectedPublicKey.PublicKeyBase58Check] = true
						affectedPublicKeyEntry := &PGAffectedPublicKeyEntry{
							AffectedPublicKeyEntry: AffectedPublicKeyEntry{
								PublicKey:       affectedPublicKey.PublicKeyBase58Check,
								Metadata:        affectedPublicKey.Metadata,
								Timestamp:       transactions[jj].Timestamp,
								TransactionHash: transactions[jj].TransactionHash,
							},
						}
						affectedPublicKeys = append(affectedPublicKeys, affectedPublicKeyEntry)
					}

					fmt.Printf("Here's transaction %d: %+v\n", jj, transactions[jj])

					transactionUpdates = append(transactionUpdates, transactions[jj])
				}
			}(jj)
		}
	}

	if len(transactionUpdates) > 0 {
		values := db.NewValues(&transactionUpdates)

		_, err := db.NewUpdate().
			With("_data", values).
			Model((*PGTransactionEntry)(nil)).
			TableExpr("_data").
			Set("tx_index_metadata = _data.tx_index_metadata").
			Set("tx_index_basic_transfer_metadata = _data.tx_index_basic_transfer_metadata").
			// Add Set for all the fields that you need to update.
			Where("pg_transaction_entry.transaction_hash = _data.transaction_hash").
			Exec(context.Background())
		if err != nil {
			return errors.Wrapf(err, "InsertTransactionEntryUtxoOps: Problem updating transactionEntryUtxoOps")
		}
	}

	// Insert affected public keys into db
	if len(affectedPublicKeys) > 0 {
		_, err := db.NewInsert().Model(&affectedPublicKeys).On("CONFLICT (public_key, transaction_hash) DO UPDATE").Exec(context.Background())
		if err != nil {
			return errors.Wrapf(err, "InsertAffectedPublicKeys: Problem inserting affectedPublicKeys")
		}
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of utxo_operation entries from the database.
func bulkDeleteUtxoOperationEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGUtxoOperationEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteUtxoOperationEntry: Error deleting entries")
	}

	return nil
}
