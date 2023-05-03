package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGDerivedKeyEntry struct {
	bun.BaseModel    `bun:"table:derived_key_entry"`
	OwnerPublicKey   string `pg:",use_zero"`
	DerivedPublicKey string `pg:",use_zero"`
	ExpirationBlock  uint64 `pg:",use_zero"`
	OperationType    uint8  `pg:",use_zero"`

	GlobalDESOLimit               uint64
	IsUnlimited                   bool
	TransactionSpendingLimitBytes []byte            `pg:",pk,use_zero"`
	ExtraData                     map[string]string `bun:"type:jsonb"`
	BadgerKey                     []byte            `pg:",pk,use_zero"`
}

// Convert the derived key DeSo encoder to the PG struct used by bun.
func DerivedKeyEncoderToPGStruct(derivedKeyEntry *lib.DerivedKeyEntry, keyBytes []byte) (*PGDerivedKeyEntry, error) {
	pgDerivedKeyEntry := &PGDerivedKeyEntry{
		OwnerPublicKey:   consumer.PublicKeyBytesToBase58Check(derivedKeyEntry.OwnerPublicKey[:]),
		DerivedPublicKey: consumer.PublicKeyBytesToBase58Check(derivedKeyEntry.DerivedPublicKey[:]),
		ExpirationBlock:  derivedKeyEntry.ExpirationBlock,
		OperationType:    uint8(derivedKeyEntry.OperationType),

		ExtraData: consumer.ExtraDataBytesToString(derivedKeyEntry.ExtraData),
		BadgerKey: keyBytes,
	}

	if derivedKeyEntry.TransactionSpendingLimitTracker != nil {
		pgDerivedKeyEntry.GlobalDESOLimit = derivedKeyEntry.TransactionSpendingLimitTracker.GlobalDESOLimit
		pgDerivedKeyEntry.IsUnlimited = derivedKeyEntry.TransactionSpendingLimitTracker.IsUnlimited
		// TODO: Figure out how to get block height in here.
		if tslBytes, err := derivedKeyEntry.TransactionSpendingLimitTracker.ToBytes(0); err != nil {
			return nil, err
		} else {
			pgDerivedKeyEntry.TransactionSpendingLimitBytes = tslBytes
		}
	}

	return pgDerivedKeyEntry, nil
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func DerivedKeyBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteDerivedKeyEntry(entries, db, operationType)
	} else {
		err = bulkInsertDerivedKeyEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertDerivedKeyEntry inserts a batch of derived_key entries into the database.
func bulkInsertDerivedKeyEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGDerivedKeyEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		if pgEntry, err := DerivedKeyEncoderToPGStruct(entry.Encoder.(*lib.DerivedKeyEntry), entry.KeyBytes); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertDerivedKeyEntry: Problem converting entry to PGEntry")
		} else {
			pgEntrySlice[ii] = pgEntry
		}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertDerivedKeyEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of derived_key entries from the database.
func bulkDeleteDerivedKeyEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGDerivedKeyEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteDerivedKeyEntry: Error deleting entries")
	}

	return nil
}