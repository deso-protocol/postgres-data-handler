package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/entries"
	"github.com/deso-protocol/postgres-data-handler/migrations/post_sync_migrations"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// PostgresDataHandler is a struct that implements the StateSyncerDataHandler interface. It is used by the
// consumer to insert/delete entries into the postgres database.
type PostgresDataHandler struct {
	// A Postgres DB used for the storage of chain data.
	DB *bun.DB
	// A bun transaction used for executing multiple operations in a single transaction.
	Txn *bun.Tx
	// Params is a struct containing the current blockchain parameters.
	// It is used to determine which prefix to use for public keys.
	Params *lib.DeSoParams
}

// HandleEntryBatch performs a bulk operation for a batch of entries, based on the encoder type.
func (postgresDataHandler *PostgresDataHandler) HandleEntryBatch(batchedEntries []*lib.StateChangeEntry) error {
	if len(batchedEntries) == 0 {
		return fmt.Errorf("PostgresDataHandler.HandleEntryBatch: No entries currently batched.")
	}

	// All entries in a batch should have the same encoder type.
	encoderType := batchedEntries[0].EncoderType

	var err error

	// Get the correct db handle.
	dbHandle := postgresDataHandler.GetDbHandle()
	// Create a savepoint in the current transaction, if the transaction exists.
	savepointName, err := postgresDataHandler.CreateSavepoint()
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.HandleEntryBatch: Error creating savepoint")
	}

	switch encoderType {
	case lib.EncoderTypePostEntry:
		err = entries.PostBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeProfileEntry:
		err = entries.ProfileBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeLikeEntry:
		err = entries.LikeBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeDiamondEntry:
		err = entries.DiamondBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeFollowEntry:
		err = entries.FollowBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeMessageEntry:
		err = entries.MessageBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeBalanceEntry:
		err = entries.BalanceBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeNFTEntry:
		err = entries.NftBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeNFTBidEntry:
		err = entries.NftBidBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeDerivedKeyEntry:
		err = entries.DerivedKeyBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeAccessGroupEntry:
		err = entries.AccessGroupBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeAccessGroupMemberEntry:
		err = entries.AccessGroupMemberBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeNewMessageEntry:
		err = entries.NewMessageBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeUserAssociationEntry:
		err = entries.UserAssociationBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypePostAssociationEntry:
		err = entries.PostAssociationBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypePKIDEntry:
		err = entries.PkidBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeDeSoBalanceEntry:
		err = entries.DesoBalanceBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeDAOCoinLimitOrderEntry:
		err = entries.DaoCoinLimitOrderBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeUtxoOperationBundle:
		err = entries.UtxoOperationBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeBlock:
		err = entries.BlockBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeTxn:
		err = entries.TransactionBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	}

	if err != nil {
		// If an error occurs, revert to the savepoint and return the error.
		rollbackErr := postgresDataHandler.RevertToSavepoint(savepointName)
		if rollbackErr != nil {
			return errors.Wrapf(rollbackErr, "PostgresDataHandler.HandleEntryBatch: Error reverting to savepoint")
		}
		return errors.Wrapf(err, "PostgresDataHandler.CallBatchOperationForEncoderType")
	}

	// Release the savepoint.
	err = postgresDataHandler.ReleaseSavepoint(savepointName)
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.HandleEntryBatch: Error releasing savepoint")
	}

	return nil
}

func (postgresDataHandler *PostgresDataHandler) HandleSyncEvent(syncEvent consumer.SyncEvent) error {
	switch syncEvent {
	case consumer.SyncEventStart:
		fmt.Println("Starting sync from beginning")
		RunMigrations(postgresDataHandler.DB, true, MigrationTypeInitial)
	case consumer.SyncEventHypersyncStart:
		fmt.Println("Starting hypersync")
	case consumer.SyncEventHypersyncComplete:
		fmt.Println("Hypersync complete")
	case consumer.SyncEventBlocksyncStart:
		fmt.Println("Starting blocksync")
		RunMigrations(postgresDataHandler.DB, false, MigrationTypePostHypersync)
		fmt.Printf("Starting to refresh explorer statistics\n")
		go post_sync_migrations.RefreshExplorerStatistics(postgresDataHandler.DB)
		// After hypersync, we don't need to maintain so many idle open connections.
		postgresDataHandler.DB.SetMaxIdleConns(4)
	}

	return nil
}

func (postgresDataHandler *PostgresDataHandler) InitiateTransaction() error {
	fmt.Printf("Initiating Txn\n")
	// If a transaction is already open, rollback the current transaction.
	if postgresDataHandler.Txn != nil {
		err := postgresDataHandler.Txn.Rollback()
		if err != nil {
			return errors.Wrapf(err, "PostgresDataHandler.InitiateTransaction: Error rolling back current transaction")
		}
	}
	tx, err := postgresDataHandler.DB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.InitiateTransaction: Error beginning transaction")
	}
	postgresDataHandler.Txn = &tx
	return nil
}

func (postgresDataHandler *PostgresDataHandler) CommitTransaction() error {
	fmt.Printf("Committing Txn\n")
	if postgresDataHandler.Txn == nil {
		return fmt.Errorf("PostgresDataHandler.CommitTransaction: No transaction to commit")
	}
	err := postgresDataHandler.Txn.Commit()
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.CommitTransaction: Error committing transaction")
	}
	postgresDataHandler.Txn = nil
	return nil
}

func (postgresDataHandler *PostgresDataHandler) RollbackTransaction() error {
	fmt.Printf("Rolling back Txn\n")
	if postgresDataHandler.Txn == nil {
		return fmt.Errorf("PostgresDataHandler.RollbackTransaction: No transaction to rollback")
	}
	err := postgresDataHandler.Txn.Rollback()
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.RollbackTransaction: Error rolling back transaction")
	}
	postgresDataHandler.Txn = nil
	return nil
}

// GetDbHandle returns the correct interface to use for database operations.
// If a transaction is open, it returns the transaction handle, otherwise it returns the db handle.
func (postgresDataHandler *PostgresDataHandler) GetDbHandle() bun.IDB {
	if postgresDataHandler.Txn != nil {
		return postgresDataHandler.Txn
	}
	return postgresDataHandler.DB
}

// CreateSavepoint creates a savepoint in the current transaction. If no transaction is open, it returns an empty string.
// The randomly generated savepoint name is returned if the savepoint is created successfully.
func (postgresDataHandler *PostgresDataHandler) CreateSavepoint() (string, error) {
	if postgresDataHandler.Txn == nil {
		return "", nil
	}
	savepointName := generateSavepointName()

	_, err := postgresDataHandler.Txn.Exec(fmt.Sprintf("SAVEPOINT %s", savepointName))
	if err != nil {
		return "", errors.Wrapf(err, "PostgresDataHandler.CreateSavepoint: Error creating savepoint")
	}

	return savepointName, nil
}

// RevertToSavepoint reverts the current transaction to the savepoint with the given name.
func (postgresDataHandler *PostgresDataHandler) RevertToSavepoint(savepointName string) error {
	if postgresDataHandler.Txn == nil {
		return nil
	}
	_, err := postgresDataHandler.Txn.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName))
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.RevertToSavepoint: Error reverting to savepoint")
	}
	return nil
}

// ReleaseSavepoint releases the savepoint with the given name.
func (postgresDataHandler *PostgresDataHandler) ReleaseSavepoint(savepointName string) error {
	if postgresDataHandler.Txn == nil {
		return nil
	}
	_, err := postgresDataHandler.Txn.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName))
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.ReleaseSavepoint: Error releasing savepoint")
	}
	return nil
}

func generateSavepointName() string {
	// Create a byte slice of length 8 for a 64-bit random value
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Handle error
		panic(err) // Example handling
	}
	// Convert the byte slice to a hexadecimal string
	return "savepoint_" + hex.EncodeToString(randomBytes)
}
