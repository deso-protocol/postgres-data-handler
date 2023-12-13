package handler

import (
	"PostgresDataHandler/entries"
	"PostgresDataHandler/migrations/post_sync_migrations"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// PostgresDataHandler is a struct that implements the StateSyncerDataHandler interface. It is used by the
// consumer to insert/delete entries into the postgres database.
type PostgresDataHandler struct {
	// A Postgres DB used for the storage of chain data.
	DB *bun.DB
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

	switch encoderType {
	case lib.EncoderTypePostEntry:
		err = entries.PostBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeProfileEntry:
		err = entries.ProfileBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeLikeEntry:
		err = entries.LikeBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeDiamondEntry:
		err = entries.DiamondBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeFollowEntry:
		err = entries.FollowBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeMessageEntry:
		err = entries.MessageBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeBalanceEntry:
		err = entries.BalanceBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeNFTEntry:
		err = entries.NftBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeNFTBidEntry:
		err = entries.NftBidBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeDerivedKeyEntry:
		err = entries.DerivedKeyBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeAccessGroupEntry:
		err = entries.AccessGroupBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeAccessGroupMemberEntry:
		err = entries.AccessGroupMemberBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeNewMessageEntry:
		err = entries.NewMessageBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeUserAssociationEntry:
		err = entries.UserAssociationBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypePostAssociationEntry:
		err = entries.PostAssociationBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypePKIDEntry:
		err = entries.PkidBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeDeSoBalanceEntry:
		err = entries.DesoBalanceBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeDAOCoinLimitOrderEntry:
		err = entries.DaoCoinLimitOrderBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeUtxoOperationBundle:
		err = entries.UtxoOperationBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeBlock:
		err = entries.BlockBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeTxn:
		err = entries.TransactionBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeStakeEntry:
		err = entries.StakeBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeValidatorEntry:
		err = entries.ValidatorBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeLockedStakeEntry:
		err = entries.LockedStakeBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeLockedBalanceEntry:
		err = entries.LockedBalanceEntryBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeLockupYieldCurvePoint:
		err = entries.LockupYieldCurvePointBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	case lib.EncoderTypeEpochEntry:
		err = entries.EpochEntryBatchOperation(batchedEntries, postgresDataHandler.DB, postgresDataHandler.Params)
	}

	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.CallBatchOperationForEncoderType")
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
