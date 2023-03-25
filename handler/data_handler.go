package handler

import (
	"PostgresDataHandler/entries"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/uptrace/bun"
)

// PostgresDataHandler is a struct that implements the StateSyncerDataHandler interface. It is used by the
// consumer to insert/delete entries into the postgres database.
type PostgresDataHandler struct {
	// A Postgres DB used for the storage of chain data.
	DB *bun.DB
}

// HandleEntry handles a single entry by inserting it into the database.
// PostgresDataHandler uses HandleEntryBatch to handle entries in bulk instead.
func (postgresDataHandler *PostgresDataHandler) HandleEntry(key []byte, encoder lib.DeSoEncoder, encoderType lib.EncoderType, operationType lib.StateSyncerOperationType) error {
	return nil
}

// HandleEntryBatch performs a bulk operation for a batch of entries, based on the encoder type.
func (postgresDataHandler *PostgresDataHandler) HandleEntryBatch(batchedEntries *consumer.BatchedEntries) error {
	if batchedEntries == nil || len(batchedEntries.Entries) == 0 {
		return fmt.Errorf("No entries currently batched.")
	}

	encoderType := batchedEntries.EncoderType

	switch encoderType {
	case lib.EncoderTypePostEntry:
		return entries.PostBatchOperation(batchedEntries.Entries, postgresDataHandler.DB, batchedEntries.OperationType)
	default:
		return nil
	}
}

func (postgresDataHandler *PostgresDataHandler) HandleSyncEvent(syncEvent consumer.SyncEvent) error {
	switch syncEvent {
	case consumer.SyncEventStart:
		RunMigrations(postgresDataHandler.DB, true, MigrationTypeInitial)
	case consumer.SyncEventHypersyncComplete:
		//TODO: Run FK migrations?
	case consumer.SyncEventComplete:
		fmt.Printf("\n***** Sync complete *****\n\n")
	}

	return nil
}
