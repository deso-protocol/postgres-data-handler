package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGPkidEntry struct {
	bun.BaseModel `bun:"table:pkid_entry"`
	Pkid          string `pg:",use_zero"`
	PublicKey     string `pg:",use_zero"`
	BadgerKey     []byte `pg:",pk,use_zero"`
}

// Convert the Diamond DeSo encoder to the PG struct used by bun.
func PkidEncoderToPGStruct(pkidEntry *lib.PKIDEntry, keyBytes []byte) *PGPkidEntry {
	return &PGPkidEntry{
		Pkid:      consumer.PublicKeyBytesToBase58Check(pkidEntry.PKID[:]),
		PublicKey: consumer.PublicKeyBytesToBase58Check(pkidEntry.PublicKey[:]),
		BadgerKey: keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func PkidBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeletePkidEntry(entries, db, operationType)
	} else {
		err = bulkInsertPkidEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertDiamondEntry inserts a batch of diamond entries into the database.
func bulkInsertPkidEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGPkidEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = PkidEncoderToPGStruct(entry.Encoder.(*lib.PKIDEntry), entry.KeyBytes)
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertPkidEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of diamond entries from the database.
func bulkDeletePkidEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGPkidEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeletePkidEntry: Error deleting entries")
	}

	return nil
}
