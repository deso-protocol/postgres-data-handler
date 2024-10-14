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
	"github.com/golang/glog"
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
		err = entries.PkidEntryBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
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
	case lib.EncoderTypeStakeEntry:
		err = entries.StakeBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeValidatorEntry:
		err = entries.ValidatorBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeLockedStakeEntry:
		err = entries.LockedStakeBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeLockedBalanceEntry:
		err = entries.LockedBalanceEntryBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeLockupYieldCurvePoint:
		err = entries.LockupYieldCurvePointBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeEpochEntry:
		err = entries.EpochEntryBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypePKID:
		err = entries.PkidBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeGlobalParamsEntry:
		err = entries.GlobalParamsBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeBLSPublicKeyPKIDPairEntry:
		err = entries.BLSPublicKeyPKIDPairBatchOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
	case lib.EncoderTypeBlockNode:
		err = entries.BlockNodeOperation(batchedEntries, dbHandle, postgresDataHandler.Params)
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
		err := postgresDataHandler.ResetAndMigrateDatabase()
		if err != nil {
			return errors.Wrapf(err, "PostgresDataHandler.HandleSyncEvent: Error resetting and migrating database")
		}
	case consumer.SyncEventHypersyncStart:
		fmt.Println("Starting hypersync")
	case consumer.SyncEventHypersyncComplete:
		fmt.Println("Hypersync complete")
	case consumer.SyncEventBlocksyncStart:
		fmt.Println("Starting blocksync")

		// Commit the transaction if it exists.
		commitTxn := postgresDataHandler.Txn != nil
		if commitTxn {
			err := postgresDataHandler.CommitTransaction()
			if err != nil {
				return errors.Wrapf(err, "PostgresDataHandler.HandleSyncEvent: Error committing transaction")
			}
		}

		if err := RunMigrations(postgresDataHandler.DB, false, MigrationTypePostHypersync); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		fmt.Printf("Starting to refresh explorer statistics\n")
		go post_sync_migrations.RefreshExplorerStatistics(postgresDataHandler.DB)

		// Begin a new transaction, if one was being tracked previously.
		if commitTxn {
			err := postgresDataHandler.InitiateTransaction()
			if err != nil {
				return errors.Wrapf(err, "PostgresDataHandler.HandleSyncEvent: Error initiating transaction")
			}
		}

		// After hypersync, we don't need to maintain so many idle open connections.
		postgresDataHandler.DB.SetMaxIdleConns(4)
	}

	return nil
}

func (postgresDataHandler *PostgresDataHandler) ResetAndMigrateDatabase() error {
	// Drop and recreate the schema - essentially nuke the entire db.
	if _, err := postgresDataHandler.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		return fmt.Errorf("failed to reset schema: %w", err)
	}

	// Run migrations.
	if err := RunMigrations(postgresDataHandler.DB, false, MigrationTypeInitial); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (postgresDataHandler *PostgresDataHandler) InitiateTransaction() error {
	// If a transaction is already open, rollback the current transaction.
	if postgresDataHandler.Txn != nil {
		if err := ReleaseAdvisoryLock(postgresDataHandler.Txn); err != nil {
			// Just log the error, but this shouldn't be a problem.
			glog.Errorf("Error releasing advisory lock: %v", err)
		}
		err := postgresDataHandler.Txn.Rollback()
		if err != nil {
			return errors.Wrapf(err, "PostgresDataHandler.InitiateTransaction: Error rolling back current transaction")
		}
	}
	if err := AcquireAdvisoryLock(postgresDataHandler.DB); err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.InitiateTransaction: Error acquiring advisory lock")
	}
	tx, err := postgresDataHandler.DB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.InitiateTransaction: Error beginning transaction")
	}
	postgresDataHandler.Txn = &tx
	return nil
}

func (postgresDataHandler *PostgresDataHandler) CommitTransaction() error {
	if postgresDataHandler.Txn == nil {
		return fmt.Errorf("PostgresDataHandler.CommitTransaction: No transaction to commit")
	}
	if err := ReleaseAdvisoryLock(postgresDataHandler.Txn); err != nil {
		// Just log the error, but this shouldn't be a problem.
		glog.Errorf("Error releasing advisory lock: %v", err)
	}
	err := postgresDataHandler.Txn.Commit()
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.CommitTransaction: Error committing transaction")
	}
	postgresDataHandler.Txn = nil
	return nil
}

func (postgresDataHandler *PostgresDataHandler) RollbackTransaction() error {
	glog.V(2).Info("Rolling back Txn\n")
	if err := ReleaseAdvisoryLock(postgresDataHandler.Txn); err != nil {
		// Just log the error, but this shouldn't be a problem.
		glog.Errorf("Error releasing advisory lock: %v", err)
	}
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

func (postgresDataHandler *PostgresDataHandler) GetParams() *lib.DeSoParams {
	return postgresDataHandler.Params
}

// GetDbHandle returns the correct interface to use for database operations.
// If a transaction is open, it returns the transaction handle, otherwise it returns the db handle.
func (postgresDataHandler *PostgresDataHandler) GetDbHandle() bun.IDB {
	if postgresDataHandler.Txn != nil {
		return postgresDataHandler.Txn
	}
	return postgresDataHandler.DB
}

func AcquireAdvisoryLock(db bun.IDB) error {
	_, err := db.NewRaw("SELECT pg_advisory_lock(1);").Exec(context.Background())
	if err != nil {
		return errors.Wrapf(err, "AcquireAdvisoryLock: Error acquiring advisory lock")
	}
	return nil
}

func ReleaseAdvisoryLock(db bun.IDB) error {
	_, err := db.NewRaw("SELECT pg_advisory_unlock(1);").Exec(context.Background())
	if err != nil {
		return errors.Wrapf(err, "ReleaseAdvisoryLock: Error releasing advisory lock")
	}
	return nil
}

// CreateSavepoint creates a savepoint in the current transaction. If no transaction is open, it returns an empty string.
// The randomly generated savepoint name is returned if the savepoint is created successfully.
func (postgresDataHandler *PostgresDataHandler) CreateSavepoint() (string, error) {
	if postgresDataHandler.Txn == nil {
		return "", nil
	}
	savepointName, err := generateSavepointName()
	if err != nil {
		return "", errors.Wrapf(err, "PostgresDataHandler.CreateSavepoint: Error generating savepoint name")
	}

	_, err = postgresDataHandler.Txn.Exec(fmt.Sprintf("SAVEPOINT %s", savepointName))
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

func generateSavepointName() (string, error) {
	// Create a byte slice of length 8 for a 64-bit random value
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.Wrapf(err, "generateSavepointName: Error generating random bytes")
	}
	// Convert the byte slice to a hexadecimal string
	return "savepoint_" + hex.EncodeToString(randomBytes), nil
}
