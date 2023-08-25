package entries

import (
	"bytes"
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
)

type BalanceEntry struct {
	HodlerPkid  string `pg:",use_zero"`
	CreatorPkid string `pg:",use_zero"`
	// Use bunbig.Int to store the balance as a numeric in the pg database.
	BalanceNanos *bunbig.Int `pg:",use_zero"`
	HasPurchased bool        `pg:",use_zero"`
	IsDaoCoin    bool        `pg:",use_zero"`
	BadgerKey    []byte      `pg:",pk,use_zero"`
}

type PGBalanceEntry struct {
	bun.BaseModel `bun:"table:balance_entry"`
	BalanceEntry
}

type PGBalanceEntryUtxoOps struct {
	bun.BaseModel `bun:"table:balance_entry_utxo_ops"`
	BalanceEntry
	UtxoOperation
}

// Convert the DeSo encoder to the postgres struct used by bun.
func BalanceEntryEncoderToPGStruct(balanceEntry *lib.BalanceEntry, keyBytes []byte, params *lib.DeSoParams) BalanceEntry {
	return BalanceEntry{
		HodlerPkid:   consumer.PublicKeyBytesToBase58Check(balanceEntry.HODLerPKID[:], params),
		CreatorPkid:  consumer.PublicKeyBytesToBase58Check(balanceEntry.CreatorPKID[:], params),
		BalanceNanos: bunbig.FromMathBig(balanceEntry.BalanceNanos.ToBig()),
		HasPurchased: balanceEntry.HasPurchased,
		// Check to see if the key has the prefix for a DAO coin.
		IsDaoCoin: bytes.Equal(keyBytes[:1], lib.Prefixes.PrefixHODLerPKIDCreatorPKIDToDAOCoinBalanceEntry),
		BadgerKey: keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate methods
// based on the operation type and executes it.
func BalanceBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteBalanceEntry(entries, db, operationType)
	} else {
		err = bulkInsertBalanceEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertBalanceEntry inserts a batch of balance entries into the database.
func bulkInsertBalanceEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGBalanceEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGBalanceEntry{BalanceEntry: BalanceEntryEncoderToPGStruct(entry.Encoder.(*lib.BalanceEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBalanceEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of balance entries from the database.
func bulkDeleteBalanceEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGBalanceEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBalanceEntry: Error deleting entries")
	}

	return nil
}
