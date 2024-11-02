package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type JailedHistoryEntry struct {
	ValidatorPKID         string `bun:",pk,nullzero"`
	JailedAtEpochNumber   uint64 `bun:",pk"`
	UnjailedAtEpochNumber uint64 `bun:",pk"`
}

type PGJailedHistoryEvent struct {
	bun.BaseModel `bun:"table:jailed_history_event"`
	JailedHistoryEntry
}

// Convert the UnjailValidatorStateChangeMetadata DeSo encoder to the JailedHistoryEntry struct used by bun.
func UnjailValidatorStateChangeMetadataEncoderToPGStruct(
	unjailValidatorStateChangeMetadata *lib.UnjailValidatorStateChangeMetadata,
	params *lib.DeSoParams,
) JailedHistoryEntry {
	pgJailedHistoryEntry := JailedHistoryEntry{
		JailedAtEpochNumber:   unjailValidatorStateChangeMetadata.JailedAtEpochNumber,
		UnjailedAtEpochNumber: unjailValidatorStateChangeMetadata.UnjailedAtEpochNumber,
	}

	if unjailValidatorStateChangeMetadata.ValidatorPKID != nil {
		pgJailedHistoryEntry.ValidatorPKID = consumer.PublicKeyBytesToBase58Check(
			(*unjailValidatorStateChangeMetadata.ValidatorPKID)[:], params)
	}

	return pgJailedHistoryEntry
}

// ValidatorBatchOperation is the entry point for processing a batch of Validator entries.
// It determines the appropriate handler based on the operation type and executes it.
func JailedHistoryEventBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams, cachedEntries *lru.Cache[string, []byte]) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteValidatorEntry(entries, db, operationType, cachedEntries)
	} else {
		err = bulkInsertValidatorEntry(entries, db, operationType, params, cachedEntries)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.ValidatorBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertJailedHistoryEvent inserts a batch of jailed history events into the database.
func bulkInsertJailedHistoryEvent(
	entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams,
) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGJailedHistoryEvent, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGJailedHistoryEvent{
			JailedHistoryEntry: UnjailValidatorStateChangeMetadataEncoderToPGStruct(
				entry.Encoder.(*lib.UnjailValidatorStateChangeMetadata), params,
			)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (validator_pkid, jailed_at_epoch_number, unjailed_at_epoch_number) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertJailedHistoryEvent: Error inserting entries")
	}
	return nil
}

// bulkDeleteJailedHistoryEvent deletes a batch of validator entries from the database.
func bulkDeleteJailedHistoryEvent(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform entries into PGJailedHistoryEvent.
	pgEntrySlice := make([]*PGJailedHistoryEvent, len(uniqueEntries))
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGJailedHistoryEvent{
			JailedHistoryEntry: UnjailValidatorStateChangeMetadataEncoderToPGStruct(
				entry.Encoder.(*lib.UnjailValidatorStateChangeMetadata), nil,
			)}
	}

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(uniqueEntries).
		WherePK().
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteValidatorEntry: Error deleting entries")
	}

	return nil
}
