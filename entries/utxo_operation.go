package entries

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
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
	IsDuplicate     bool      `pg:",pk,use_zero"`
	Timestamp       time.Time `pg:",use_zero"`
	TxnType         uint16    `pg:",use_zero"`
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

// UtxoOperationBatchOperation is the entry point for processing a batch of utxo operations. It determines the appropriate handler
// based on the operation type and executes it.
func UtxoOperationBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteUtxoOperationEntry(entries, db, operationType)
	} else {
		err = bulkInsertUtxoOperationsEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertUtxoOperationsEntry inserts a batch of utxo operation entries into the database.
func bulkInsertUtxoOperationsEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {

	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transactions added to this slice will have their txindex metadata updated.
	transactionUpdates := make([]*PGTransactionEntry, 0)
	affectedPublicKeys := make([]*PGAffectedPublicKeyEntry, 0)
	blockEntries := make([]*PGBlockEntry, 0)
	stakeRewardEntries := make([]*PGStakeReward, 0)
	jailedHistoryEntries := make([]*PGJailedHistoryEvent, 0)

	// Start timer to track how long it takes to insert the entries.
	start := time.Now()

	fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Inserting %v entries\n", len(uniqueEntries))
	transactionCount := 0

	// Whether we are inserting transactions for the first time, or just updating them.
	// On initial sync it will be inserting, otherwise it will be a bulk update.
	insertTransactions := false

	// Loop through the utxo op bundles and extract the utxo operation entries from them.
	for _, entry := range uniqueEntries {

		transactions := []*PGTransactionEntry{}

		// We can use this function regardless of the db prefix, because both block_hash and transaction_hash
		// are stored in the same blockHashHex format in the key.
		blockHash := ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes)

		// Check to see if the state change entry has an attached block.
		// Note that this only happens during the initial sync, in order to speed up the sync process.
		if entry.Block != nil {
			insertTransactions = true
			block := entry.Block
			blockEntry := BlockEncoderToPGStruct(block, entry.KeyBytes, params)
			blockEntries = append(blockEntries, blockEntry)
			for ii, txn := range block.Txns {
				// Check if the transaction connects or not.
				txnConnects := params.IsPoWBlockHeight(blockEntry.Height) ||
					ii == 0 || block.TxnConnectStatusByIndex.Get(ii-1)
				pgTxn, err := TransactionEncoderToPGStruct(
					txn, uint64(ii), blockEntry.BlockHash, blockEntry.Height, blockEntry.Timestamp, txnConnects, params)
				if err != nil {
					return errors.Wrapf(err, "entries.bulkInsertUtxoOperationsEntry: Problem converting transaction to PG struct")
				}
				transactions = append(transactions, pgTxn)
			}
		} else {
			// If the block isn't available on the entry itself, we can retrieve it from the database.
			filterField := ""

			// Check if the entry is a bundle with multiple utxo operations, or a single transaction.
			// If bundle, get a list of transactions based on the block hash extracted from the key.
			// If single transaction, get the transaction based on the transaction hash, extracted from the key.

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
			err := db.NewSelect().Model(&transactions).Column("txn_bytes", "transaction_hash", "timestamp", "txn_type").Where(fmt.Sprintf("%s = ?", filterField), blockHash).Order("index_in_block ASC").Scan(context.Background())
			if err != nil {
				return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem getting transactions at block height %v: %v", entry.BlockHeight, err)
			}
		}

		utxoOperations, ok := entry.Encoder.(*lib.UtxoOperationBundle)
		if !ok {
			return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem with entry %v", entry)
		}

		transactionCount += len(utxoOperations.UtxoOpBundle)
		// Create a wait group to wait for all the goroutines to finish.
		for jj := range utxoOperations.UtxoOpBundle {

			utxoOps := utxoOperations.UtxoOpBundle[jj]
			// Update the transaction metadata for this transaction.
			if jj < len(transactions) {
				transaction := &lib.MsgDeSoTxn{}
				err := transaction.FromBytes(transactions[jj].TxnBytes)
				if err != nil {
					return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem decoding transaction for entry %+v at block height %v", entry, entry.BlockHeight)
				}
				txIndexMetadata, err := consumer.ComputeTransactionMetadata(transaction, blockHash, params, transaction.TxnFeeNanos, uint64(jj), utxoOps)
				if err != nil {
					return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem computing transaction metadata for entry %+v at block height %v", entry, entry.BlockHeight)
				}

				metadata := txIndexMetadata.GetEncoderForTxType(transaction.TxnMeta.GetTxnType())
				basicTransferMetadata := txIndexMetadata.BasicTransferTxindexMetadata
				basicTransferMetadata.UtxoOps = nil

				transactions[jj].TxIndexMetadata = metadata

				transactions[jj].TxIndexBasicTransferMetadata = txIndexMetadata.GetEncoderForTxType(lib.TxnTypeBasicTransfer)

				// Track which public keys have already been added to the affected public keys slice, to avoid duplicates.
				affectedPublicKeyMetadataSet := make(map[string]bool)
				affectedPublicKeySet := make(map[string]bool)

				switch transaction.TxnMeta.GetTxnType() {
				case lib.TxnTypeUnjailValidator:
					// Find the unjail utxo op
					var unjailUtxoOp *lib.UtxoOperation
					for _, utxoOp := range utxoOps {
						if utxoOp.Type == lib.OperationTypeUnjailValidator {
							unjailUtxoOp = utxoOp
							break
						}
					}
					if unjailUtxoOp == nil {
						glog.Error("bulkInsertUtxoOperationsEntry: Problem finding unjail utxo op")
						continue
					}
					scm, ok := unjailUtxoOp.StateChangeMetadata.(*lib.UnjailValidatorStateChangeMetadata)
					if !ok {
						glog.Error("bulkInsertUtxoOperationsEntry: Problem with state change metadata for unjail")
						continue
					}
					// Parse the jailed history event and add it to the slice.
					jailedHistoryEntries = append(jailedHistoryEntries,
						&PGJailedHistoryEvent{
							JailedHistoryEntry: UnjailValidatorStateChangeMetadataEncoderToPGStruct(scm, params),
						},
					)
				}

				// Loop through the affected public keys and add them to the affected public keys slice.
				for _, affectedPublicKey := range txIndexMetadata.AffectedPublicKeys {
					// Skip if we've already added this public key/metadata.
					apkmDuplicateKey := fmt.Sprintf("%v:%v", affectedPublicKey.PublicKeyBase58Check, affectedPublicKey.Metadata)
					if _, ok := affectedPublicKeyMetadataSet[apkmDuplicateKey]; ok {
						continue
					}
					affectedPublicKeyMetadataSet[apkmDuplicateKey] = true

					// Track which public keys have already been added to the affected public keys slice. If they have,
					// mark this record as a duplicate to make it easier to filter out.
					apkIsDuplicate := false
					if _, ok := affectedPublicKeySet[affectedPublicKey.PublicKeyBase58Check]; ok {
						apkIsDuplicate = true
					}
					affectedPublicKeySet[affectedPublicKey.PublicKeyBase58Check] = true

					affectedPublicKeyEntry := &PGAffectedPublicKeyEntry{
						AffectedPublicKeyEntry: AffectedPublicKeyEntry{
							PublicKey:       affectedPublicKey.PublicKeyBase58Check,
							Metadata:        affectedPublicKey.Metadata,
							IsDuplicate:     apkIsDuplicate,
							Timestamp:       transactions[jj].Timestamp,
							TxnType:         transactions[jj].TxnType,
							TransactionHash: transactions[jj].TransactionHash,
						},
					}
					affectedPublicKeys = append(affectedPublicKeys, affectedPublicKeyEntry)
				}
				transactionUpdates = append(transactionUpdates, transactions[jj])
			} else if jj == len(transactions) {
				// TODO: parse utxo operations for the block level index.
				// Examples: deletion of expired nonces, staking rewards (restaked
				// + payed to balance), validator jailing, updating validator's
				// last active at epoch.
				for ii, utxoOp := range utxoOps {
					switch utxoOp.Type {
					case lib.OperationTypeStakeDistributionRestake, lib.OperationTypeStakeDistributionPayToBalance:
						stateChangeMetadata, ok := utxoOp.StateChangeMetadata.(*lib.StakeRewardStateChangeMetadata)
						if !ok {
							glog.Error("bulkInsertUtxoOperationsEntry: Problem with state change metadata for " +
								"stake rewards")
							continue
						}
						stakeReward := PGStakeReward{
							StakeReward: StakeRewardEncoderToPGStruct(stateChangeMetadata, params, blockHash, uint64(ii)),
						}
						stakeRewardEntries = append(stakeRewardEntries, &stakeReward)
					}
				}

			}
		}
		// Print how long it took to insert the entries.
	}
	fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Processed %v txns in %v s\n", transactionCount, time.Since(start))

	start = time.Now()

	if len(transactionUpdates) > 0 {

		if insertTransactions {
			err := bulkInsertTransactionEntry(transactionUpdates, db, operationType)
			if err != nil {
				return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem inserting transaction entries: %v", err)
			}

			blockQuery := db.NewInsert().Model(&blockEntries)

			if operationType == lib.DbOperationTypeUpsert {
				blockQuery = blockQuery.On("CONFLICT (block_hash) DO UPDATE")
			}

			if _, err := blockQuery.Exec(context.Background()); err != nil {
				return errors.Wrapf(err, "entries.bulkInsertBlock: Error inserting entries")
			}

		} else {
			values := db.NewValues(&transactionUpdates)
			_, err := db.NewUpdate().
				With("_data", values).
				Model((*PGTransactionEntry)(nil)).
				TableExpr("_data").
				Set("tx_index_metadata = _data.tx_index_metadata").
				Set("tx_index_basic_transfer_metadata = _data.tx_index_basic_transfer_metadata").
				// Add Set for all the fields that you need to update.
				Where("pg_transaction_entry.transaction_hash = _data.transaction_hash").
				Where("pg_transaction_entry.txn_type = _data.txn_type").
				Exec(context.Background())
			if err != nil {
				return errors.Wrapf(err, "InsertTransactionEntryUtxoOps: Problem updating transactionEntryUtxoOps")
			}
		}
	}

	fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Updated %v txns in %v s\n", len(transactionUpdates), time.Since(start))

	start = time.Now()

	// Insert affected public keys into db
	if len(affectedPublicKeys) > 0 {
		_, err := db.NewInsert().Model(&affectedPublicKeys).On("CONFLICT (public_key, transaction_hash, metadata) DO UPDATE").Exec(context.Background())
		if err != nil {
			return errors.Wrapf(err, "InsertAffectedPublicKeys: Problem inserting affectedPublicKeys")
		}
	}

	fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Inserted %v affected public keys in %v s\n", len(affectedPublicKeys), time.Since(start))

	start = time.Now()

	// Insert stake rewards into db
	if len(stakeRewardEntries) > 0 {
		_, err := db.NewInsert().Model(&stakeRewardEntries).On("CONFLICT (block_hash, utxo_op_index) DO UPDATE").Exec(context.Background())
		if err != nil {
			return errors.Wrapf(err, "InsertStakeRewards: Problem inserting stake rewards")
		}
	}
	fmt.Printf("entries.bulkInsertUtxoOperationsEntry: Inserted %v stake rewards in %v s\n", len(stakeRewardEntries), time.Since(start))

	if len(jailedHistoryEntries) > 0 {
		_, err := db.NewInsert().Model(&jailedHistoryEntries).On("CONFLICT (validator_pkid, jailed_at_epoch_number, unjailed_at_epoch_number) DO NOTHING").Exec(context.Background())
		if err != nil {
			return errors.Wrapf(err, "InsertJailedHistory: Problem inserting jailed history")
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
