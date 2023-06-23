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
	PublicKey       string `pg:",pk,use_zero"`
	TransactionHash string `pg:",pk,use_zero"`
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
func UtxoOperationBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
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

// bulkInsertUtxoOperationsEntry inserts a batch of user_association entries into the database.
func bulkInsertUtxoOperationsEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {

	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	//pgEntrySlice := make([]*PGUtxoOperationEntry, len(uniqueEntries))

	//postEntryUtxoOps := make([]*PGPostEntryUtxoOps, 0)
	//profileEntryUtxoOps := make([]*PGProfileEntryUtxoOps, 0)
	//likeEntryUtxoOps := make([]*PGLikeEntryUtxoOps, 0)
	//diamondEntryUtxoOps := make([]*PGDiamondEntryUtxoOps, 0)
	//nftEntryUtxoOps := make([]*PGNftEntryUtxoOps, 0)
	//nftBidEntryUtxoOps := make([]*PGNftBidEntryUtxoOps, 0)
	//derivedKeyEntryUtxoOps := make([]*PGDerivedKeyEntryUtxoOps, 0)
	//newMessageEntryUtxoOps := make([]*PGNewMessageEntryUtxoOps, 0)
	//balanceEntryUtxoOps := make([]*PGBalanceEntryUtxoOps, 0)
	//userAssociationEntryUtxoOps := make([]*PGUserAssociationEntryUtxoOps, 0)
	//postAssociationEntryUtxoOps := make([]*PGPostAssociationEntryUtxoOps, 0)
	//accessGroupEntryUtxoOps := make([]*PGAccessGroupEntryUtxoOps, 0)
	//accessGroupMemberEntryUtxoOps := make([]*PGAccessGroupMemberEntryUtxoOps, 0)
	//desoBalanceEntryUtxoOps := make([]*PGDesoBalanceEntryUtxoOps, 0)

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

		for jj, utxoOps := range utxoOperations.UtxoOpBundle {

			// Update the transaction metadata for this transaction.
			if jj < len(transactions) {
				transaction := &lib.MsgDeSoTxn{}
				err = transaction.FromBytes(transactions[jj].TxnBytes)
				if err != nil {
					return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem decoding transaction for entry %+v at block height %v", entry, entry.BlockHeight)
				}
				txIndexMetadata, err := consumer.ComputeTransactionMetadata(transaction, blockHash, &lib.DeSoMainnetParams, transaction.TxnFeeNanos, uint64(jj), utxoOps)
				if err != nil {
					return fmt.Errorf("entries.bulkInsertUtxoOperationsEntry: Problem computing transaction metadata for entry %+v at block height %v", entry, entry.BlockHeight)
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
							TransactionHash: transactions[jj].TransactionHash,
						},
					}
					affectedPublicKeys = append(affectedPublicKeys, affectedPublicKeyEntry)
				}

				transactionUpdates = append(transactionUpdates, transactions[jj])
			}

			//for utxoOpIndex, utxoOp := range utxoOps {
			//	pgEntrySlice[ii] = UtxoOperationEncoderToPGStruct(utxoOp, entry.KeyBytes, uint64(jj), uint64(utxoOpIndex), entry.BlockHeight)
			//	//	//			// TODO: Reduce duplicated code here.
			//	//	if utxoOp.PrevPostEntry != nil {
			//	//		postEntry, _ := PostEntryEncoderToPGStruct(utxoOp.PrevPostEntry, []byte{})
			//	//		prevPostEntry := &PGPostEntryUtxoOps{
			//	//			PostEntry: postEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevPostEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		postEntryUtxoOps = append(postEntryUtxoOps, prevPostEntry)
			//	//	}
			//	//	if utxoOp.PrevParentPostEntry != nil {
			//	//		postEntry, _ := PostEntryEncoderToPGStruct(utxoOp.PrevParentPostEntry, []byte{})
			//	//		prevPostEntry := &PGPostEntryUtxoOps{
			//	//			PostEntry: postEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevParentPostEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		postEntryUtxoOps = append(postEntryUtxoOps, prevPostEntry)
			//	//	}
			//	//	if utxoOp.PrevGrandparentPostEntry != nil {
			//	//		postEntry, _ := PostEntryEncoderToPGStruct(utxoOp.PrevGrandparentPostEntry, []byte{})
			//	//		prevPostEntry := &PGPostEntryUtxoOps{
			//	//			PostEntry: postEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevGrandparentPostEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		postEntryUtxoOps = append(postEntryUtxoOps, prevPostEntry)
			//	//	}
			//	//	if utxoOp.PrevRepostedPostEntry != nil {
			//	//		postEntry, _ := PostEntryEncoderToPGStruct(utxoOp.PrevRepostedPostEntry, []byte{})
			//	//		prevPostEntry := &PGPostEntryUtxoOps{
			//	//			PostEntry: postEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevRepostedPostEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		postEntryUtxoOps = append(postEntryUtxoOps, prevPostEntry)
			//	//	}
			//	//	if utxoOp.PrevProfileEntry != nil {
			//	//		profileEntry := ProfileEntryEncoderToPGStruct(utxoOp.PrevProfileEntry, append(lib.Prefixes.PrefixPKIDToProfileEntry, utxoOp.PrevProfileEntry.PublicKey...))
			//	//		prevProfileEntry := &PGProfileEntryUtxoOps{
			//	//			ProfileEntry: profileEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevProfileEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		profileEntryUtxoOps = append(profileEntryUtxoOps, prevProfileEntry)
			//	//	}
			//	//	if utxoOp.PrevLikeEntry != nil {
			//	//		likeEntry := LikeEncoderToPGStruct(utxoOp.PrevLikeEntry, []byte{})
			//	//		prevLikeEntry := &PGLikeEntryUtxoOps{
			//	//			LikeEntry: likeEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevLikeEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		likeEntryUtxoOps = append(likeEntryUtxoOps, prevLikeEntry)
			//	//	}
			//	//	if utxoOp.PrevDiamondEntry != nil {
			//	//		diamondEntry := DiamondEncoderToPGStruct(utxoOp.PrevDiamondEntry, []byte{})
			//	//		prevDiamondEntry := &PGDiamondEntryUtxoOps{
			//	//			DiamondEntry: diamondEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevDiamondEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		diamondEntryUtxoOps = append(diamondEntryUtxoOps, prevDiamondEntry)
			//	//	}
			//	//	if utxoOp.PrevNFTEntry != nil {
			//	//		nftEntry := NftEncoderToPGStruct(utxoOp.PrevNFTEntry, []byte{})
			//	//		prevNFTEntry := &PGNftEntryUtxoOps{
			//	//			NftEntry: nftEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevNFTEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		nftEntryUtxoOps = append(nftEntryUtxoOps, prevNFTEntry)
			//	//	}
			//	//	if utxoOp.PrevNFTBidEntry != nil {
			//	//		nftBidEntry := NftBidEncoderToPGStruct(utxoOp.PrevNFTBidEntry, []byte{})
			//	//		prevNFTBidEntry := &PGNftBidEntryUtxoOps{
			//	//			NftBidEntry: nftBidEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevNFTBidEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		nftBidEntryUtxoOps = append(nftBidEntryUtxoOps, prevNFTBidEntry)
			//	//	}
			//	//	if utxoOp.DeletedNFTBidEntries != nil {
			//	//		for jj, nftBidEntry := range utxoOp.DeletedNFTBidEntries {
			//	//			deletedNftBidEntry := NftBidEncoderToPGStruct(nftBidEntry, []byte{})
			//	//			prevNFTBidEntry := &PGNftBidEntryUtxoOps{
			//	//				NftBidEntry: deletedNftBidEntry,
			//	//				UtxoOperation: UtxoOperation{
			//	//					UtxoOpEntryType:  "DeletedNFTBidEntry",
			//	//					UtxoOpIndex:      uint64(utxoOpIndex),
			//	//					TransactionIndex: uint64(jj),
			//	//					ArrayIndex:       uint64(jj),
			//	//					BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//				},
			//	//			}
			//	//			nftBidEntryUtxoOps = append(nftBidEntryUtxoOps, prevNFTBidEntry)
			//	//		}
			//	//	}
			//	//	if utxoOp.PrevAcceptedNFTBidEntries != nil {
			//	//		for jj, nftBidEntry := range *utxoOp.PrevAcceptedNFTBidEntries {
			//	//			acceptedNftBidEntry := NftBidEntry{}
			//	//			if nftBidEntry != nil && nftBidEntry.BidderPKID != nil {
			//	//				acceptedNftBidEntry = NftBidEncoderToPGStruct(nftBidEntry, []byte{})
			//	//			}
			//	//			prevNFTBidEntry := &PGNftBidEntryUtxoOps{
			//	//				NftBidEntry: acceptedNftBidEntry,
			//	//				UtxoOperation: UtxoOperation{
			//	//					UtxoOpEntryType:  "PrevAcceptedNFTBidEntry",
			//	//					UtxoOpIndex:      uint64(utxoOpIndex),
			//	//					TransactionIndex: uint64(jj),
			//	//					ArrayIndex:       uint64(jj),
			//	//					BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//				},
			//	//			}
			//	//			nftBidEntryUtxoOps = append(nftBidEntryUtxoOps, prevNFTBidEntry)
			//	//		}
			//	//	}
			//	//	if utxoOp.PrevDerivedKeyEntry != nil {
			//	//		derivedKeyEntry, _ := DerivedKeyEncoderToPGStruct(utxoOp.PrevDerivedKeyEntry, []byte{})
			//	//		prevDerivedKeyEntry := &PGDerivedKeyEntryUtxoOps{
			//	//			DerivedKeyEntry: derivedKeyEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevDerivedKeyEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		derivedKeyEntryUtxoOps = append(derivedKeyEntryUtxoOps, prevDerivedKeyEntry)
			//	//	}
			//	//	if utxoOp.PrevTransactorBalanceEntry != nil {
			//	//		prevTransactorBalanceEntry := BalanceEntryEncoderToPGStruct(utxoOp.PrevTransactorBalanceEntry, []byte{})
			//	//		prevTransactorBalanceEntryUtxoOps := &PGBalanceEntryUtxoOps{
			//	//			BalanceEntry: prevTransactorBalanceEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevTransactorBalanceEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		balanceEntryUtxoOps = append(balanceEntryUtxoOps, prevTransactorBalanceEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevCreatorBalanceEntry != nil {
			//	//		prevCreatorBalanceEntry := BalanceEntryEncoderToPGStruct(utxoOp.PrevCreatorBalanceEntry, []byte{})
			//	//		prevCreatorBalanceEntryUtxoOps := &PGBalanceEntryUtxoOps{
			//	//			BalanceEntry: prevCreatorBalanceEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevCreatorBalanceEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		balanceEntryUtxoOps = append(balanceEntryUtxoOps, prevCreatorBalanceEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevSenderBalanceEntry != nil {
			//	//		prevSenderBalanceEntry := BalanceEntryEncoderToPGStruct(utxoOp.PrevSenderBalanceEntry, []byte{})
			//	//		prevSenderBalanceEntryUtxoOps := &PGBalanceEntryUtxoOps{
			//	//			BalanceEntry: prevSenderBalanceEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevSenderBalanceEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		balanceEntryUtxoOps = append(balanceEntryUtxoOps, prevSenderBalanceEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevReceiverBalanceEntry != nil {
			//	//		prevReceiverBalanceEntry := BalanceEntryEncoderToPGStruct(utxoOp.PrevReceiverBalanceEntry, []byte{})
			//	//		prevReceiverBalanceEntryUtxoOps := &PGBalanceEntryUtxoOps{
			//	//			BalanceEntry: prevReceiverBalanceEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevReceiverBalanceEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		balanceEntryUtxoOps = append(balanceEntryUtxoOps, prevReceiverBalanceEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevUserAssociationEntry != nil {
			//	//		prevUserAssociationEntry := UserAssociationEncoderToPGStruct(utxoOp.PrevUserAssociationEntry, []byte{})
			//	//		prevUserAssociationEntryUtxoOps := &PGUserAssociationEntryUtxoOps{
			//	//			UserAssociationEntry: prevUserAssociationEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevUserAssociationEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		userAssociationEntryUtxoOps = append(userAssociationEntryUtxoOps, prevUserAssociationEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevPostAssociationEntry != nil {
			//	//		prevPostAssociationEntry := PostAssociationEncoderToPGStruct(utxoOp.PrevPostAssociationEntry, []byte{})
			//	//		prevPostAssociationEntryUtxoOps := &PGPostAssociationEntryUtxoOps{
			//	//			PostAssociationEntry: prevPostAssociationEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevPostAssociationEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		postAssociationEntryUtxoOps = append(postAssociationEntryUtxoOps, prevPostAssociationEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevAccessGroupEntry != nil {
			//	//		prevAccessGroupEntry := AccessGroupEncoderToPGStruct(utxoOp.PrevAccessGroupEntry, []byte{})
			//	//		prevAccessGroupEntryUtxoOps := &PGAccessGroupEntryUtxoOps{
			//	//			AccessGroupEntry: prevAccessGroupEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevAccessGroupEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		accessGroupEntryUtxoOps = append(accessGroupEntryUtxoOps, prevAccessGroupEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.PrevAccessGroupMembersList != nil && len(utxoOp.PrevAccessGroupMembersList) > 0 {
			//	//		for jj, prevAccessGroupMembersList := range utxoOp.PrevAccessGroupMembersList {
			//	//			prevAccessGroupMembersListEntry := AccessGroupMemberEncoderToPGStruct(prevAccessGroupMembersList, []byte{})
			//	//			prevAccessGroupMembersListEntryUtxoOps := &PGAccessGroupMemberEntryUtxoOps{
			//	//				AccessGroupMemberEntry: prevAccessGroupMembersListEntry,
			//	//				UtxoOperation: UtxoOperation{
			//	//					UtxoOpEntryType:  "PrevAccessGroupMembersList",
			//	//					UtxoOpIndex:      uint64(utxoOpIndex),
			//	//					TransactionIndex: uint64(jj),
			//	//					ArrayIndex:       uint64(jj),
			//	//					BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//				},
			//	//			}
			//	//			accessGroupMemberEntryUtxoOps = append(accessGroupMemberEntryUtxoOps, prevAccessGroupMembersListEntryUtxoOps)
			//	//		}
			//	//	}
			//	//	if utxoOp.PrevNewMessageEntry != nil {
			//	//		prevNewMessageEntry := NewMessageEncoderToPGStruct(utxoOp.PrevNewMessageEntry, []byte{})
			//	//		prevNewMessageEntryUtxoOps := &PGNewMessageEntryUtxoOps{
			//	//			NewMessageEntry: prevNewMessageEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevNewMessageEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		newMessageEntryUtxoOps = append(newMessageEntryUtxoOps, prevNewMessageEntryUtxoOps)
			//	//	}
			//	//	if utxoOp.BalanceAmountNanos != 0 {
			//	//		prevDesoBalanceEntry := DesoBalanceEncoderToPGStruct(&lib.DeSoBalanceEntry{
			//	//			PKID:         lib.NewPKID(utxoOp.BalancePublicKey),
			//	//			BalanceNanos: utxoOp.BalanceAmountNanos,
			//	//		}, []byte{})
			//	//		prevDesoBalanceUtxoOp := &PGDesoBalanceEntryUtxoOps{
			//	//			DesoBalanceEntry: prevDesoBalanceEntry,
			//	//			UtxoOperation: UtxoOperation{
			//	//				UtxoOpEntryType:  "PrevDesoBalanceEntry",
			//	//				UtxoOpIndex:      uint64(utxoOpIndex),
			//	//				TransactionIndex: uint64(jj),
			//	//				BlockHash:        ConvertUtxoOperationKeyToBlockHashHex(entry.KeyBytes),
			//	//			},
			//	//		}
			//	//		desoBalanceEntryUtxoOps = append(desoBalanceEntryUtxoOps, prevDesoBalanceUtxoOp)
			//	//	}
			//}
		}
	}

	// Insert the entries into the database.
	//if len(postEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&postEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertPostEntryUtxoOps: Problem inserting postEntryUtxoOps")
	//	}
	//}
	//if len(profileEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&profileEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertProfileEntryUtxoOps: Problem inserting profileEntryUtxoOps")
	//	}
	//}
	//if len(likeEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&likeEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertLikeEntryUtxoOps: Problem inserting likeEntryUtxoOps")
	//	}
	//}
	//if len(diamondEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&diamondEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertDiamondEntryUtxoOps: Problem inserting diamondEntryUtxoOps")
	//	}
	//}
	//if len(nftEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&nftEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertNftEntryUtxoOps: Problem inserting nftEntryUtxoOps")
	//	}
	//}
	//if len(nftBidEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&nftBidEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertNftBidEntryUtxoOps: Problem inserting nftBidEntryUtxoOps")
	//	}
	//}
	//if len(derivedKeyEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&derivedKeyEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertDerivedKeyEntryUtxoOps: Problem inserting derivedKeyEntryUtxoOps")
	//	}
	//}
	//if len(newMessageEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&newMessageEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertNewMessageEntryUtxoOps: Problem inserting newMessageEntryUtxoOps")
	//	}
	//}
	//if len(balanceEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&balanceEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertBalanceEntryUtxoOps: Problem inserting balanceEntryUtxoOps")
	//	}
	//}
	//if len(userAssociationEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&userAssociationEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertUserAssociationEntryUtxoOps: Problem inserting userAssociationEntryUtxoOps")
	//	}
	//}
	//if len(accessGroupEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&accessGroupEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertAccessGroupEntryUtxoOps: Problem inserting accessGroupEntryUtxoOps")
	//	}
	//}
	//if len(accessGroupMemberEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&accessGroupMemberEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertAccessGroupMemberEntryUtxoOps: Problem inserting accessGroupMemberEntryUtxoOps")
	//	}
	//}
	//if len(desoBalanceEntryUtxoOps) > 0 {
	//	if _, err := db.NewInsert().Model(&desoBalanceEntryUtxoOps).Exec(context.Background()); err != nil {
	//		return errors.Wrapf(err, "InsertDesoBalanceEntryUtxoOps: Problem inserting desoBalanceEntryUtxoOps")
	//	}
	//}

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

	// Insert utxo ops into db
	//query := db.NewInsert().Model(&pgEntrySlice)

	//if operationType == lib.DbOperationTypeUpsert {
	//	query = query.On("CONFLICT (block_hash, transaction_index, utxo_op_index) DO UPDATE")
	//}
	//
	//if _, err := query.Returning("").Exec(context.Background()); err != nil {
	//	return errors.Wrapf(err, "entries.bulkInsertUtxoOperation: Error inserting entries")
	//}

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
