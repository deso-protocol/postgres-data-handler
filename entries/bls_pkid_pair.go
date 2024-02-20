package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// TODO: when to use nullzero vs use_zero?
type BLSPublicKeyPKIDPairEntry struct {
	PKID         string `bun:",nullzero"`
	BLSPublicKey string `bun:",nullzero"`

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGBLSPkidPairEntry struct {
	bun.BaseModel `bun:"table:bls_public_key_pkid_pair_entry"`
	BLSPublicKeyPKIDPairEntry
}

// Convert the BLSPublicKeyPKIDPairEntry DeSo encoder to the PGBLSPkidPairEntry struct used by bun.
func BLSPublicKeyPKIDPairEncoderToPGStruct(blsPublicKeyPKIDPairEntry *lib.BLSPublicKeyPKIDPairEntry, keyBytes []byte, params *lib.DeSoParams) BLSPublicKeyPKIDPairEntry {
	pgBLSPkidPairEntry := BLSPublicKeyPKIDPairEntry{
		BadgerKey: keyBytes,
	}

	if blsPublicKeyPKIDPairEntry.PKID != nil {
		pgBLSPkidPairEntry.PKID = consumer.PublicKeyBytesToBase58Check((*blsPublicKeyPKIDPairEntry.PKID)[:], params)
	}

	if !blsPublicKeyPKIDPairEntry.BLSPublicKey.IsEmpty() {
		pgBLSPkidPairEntry.BLSPublicKey = blsPublicKeyPKIDPairEntry.BLSPublicKey.ToString()
	}

	return pgBLSPkidPairEntry
}

// BLSPublicKeyPKIDPairBatchOperation is the entry point for processing a batch of BLSPublicKeyPKIDPair entries.
// It determines the appropriate handler based on the operation type and executes it.
func BLSPublicKeyPKIDPairBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteBLSPkidPairEntry(entries, db, operationType)
	} else {
		err = bulkInsertBLSPkidPairEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.StakeBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertBLSPkidPairEntry inserts a batch of stake entries into the database.
func bulkInsertBLSPkidPairEntry(
	entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams,
) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGBLSPkidPairEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGBLSPkidPairEntry{BLSPublicKeyPKIDPairEntry: BLSPublicKeyPKIDPairEncoderToPGStruct(
			entry.Encoder.(*lib.BLSPublicKeyPKIDPairEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertBLSPkidPairEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteBLSPkidPairEntry deletes a batch of stake entries from the database.
func bulkDeleteBLSPkidPairEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGBLSPkidPairEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteBLSPkidPairEntry: Error deleting entries")
	}

	return nil
}
