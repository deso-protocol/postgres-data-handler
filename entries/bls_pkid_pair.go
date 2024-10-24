package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	lru "github.com/hashicorp/golang-lru/v2"
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

type BLSPublicKeyPKIDPairSnapshotEntry struct {
	PKID                  string `bun:",nullzero"`
	BLSPublicKey          string `bun:",nullzero"`
	SnapshotAtEpochNumber uint64 `pg:",use_zero"`

	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGBLSPublicKeyPKIDPairSnapshotEntry struct {
	bun.BaseModel `bun:"table:bls_public_key_pkid_pair_snapshot_entry"`
	BLSPublicKeyPKIDPairSnapshotEntry
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

// BLSPublicKeyPKIDPairSnapshotEncoderToPGStruct converts the BLSPublicKeyPKIDPairSnapshotEntry DeSo encoder to the
// PGBLSPublicKeyPKIDPairSnapshotEntry struct used by bun.
func BLSPublicKeyPKIDPairSnapshotEncoderToPGStruct(
	blsPublicKeyPKIDPairEntry *lib.BLSPublicKeyPKIDPairEntry, keyBytes []byte, params *lib.DeSoParams,
) BLSPublicKeyPKIDPairSnapshotEntry {
	prefixRemovedKeyBytes := keyBytes[1:]
	epochNumber := lib.DecodeUint64(prefixRemovedKeyBytes[:8])

	pgBLSPkidPairSnapshotEntry := BLSPublicKeyPKIDPairSnapshotEntry{
		SnapshotAtEpochNumber: epochNumber,
		BadgerKey:             keyBytes,
	}

	if blsPublicKeyPKIDPairEntry.PKID != nil {
		pgBLSPkidPairSnapshotEntry.PKID = consumer.PublicKeyBytesToBase58Check((*blsPublicKeyPKIDPairEntry.PKID)[:], params)
	}

	if !blsPublicKeyPKIDPairEntry.BLSPublicKey.IsEmpty() {
		pgBLSPkidPairSnapshotEntry.BLSPublicKey = blsPublicKeyPKIDPairEntry.BLSPublicKey.ToString()
	}

	return pgBLSPkidPairSnapshotEntry
}

// BLSPublicKeyPKIDPairBatchOperation is the entry point for processing a batch of BLSPublicKeyPKIDPair entries.
// It determines the appropriate handler based on the operation type and executes it.
func BLSPublicKeyPKIDPairBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams, cachedEntries *lru.Cache[string, []byte]) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteBLSPkidPairEntry(entries, db, operationType, cachedEntries)
	} else {
		err = bulkInsertBLSPkidPairEntry(entries, db, operationType, params, cachedEntries)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.StakeBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertBLSPkidPairEntry inserts a batch of stake entries into the database.
func bulkInsertBLSPkidPairEntry(
	entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams, cachedEntries *lru.Cache[string, []byte],
) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Filter out any entries that are already tracked in the cache.
	uniqueEntries = consumer.FilterCachedEntries(uniqueEntries, cachedEntries)

	uniqueBLSPkidPairEntries := consumer.FilterEntriesByPrefix(
		uniqueEntries, lib.Prefixes.PrefixValidatorBLSPublicKeyPKIDPairEntry)
	uniqueBLSPkidPairSnapshotEntries := consumer.FilterEntriesByPrefix(
		uniqueEntries, lib.Prefixes.PrefixSnapshotValidatorBLSPublicKeyPKIDPairEntry)
	// Create a new array to hold the bun struct.
	pgBLSPkidPairEntrySlice := make([]*PGBLSPkidPairEntry, len(uniqueBLSPkidPairEntries))
	pgBLSPkidPairSnapshotEntrySlice := make([]*PGBLSPublicKeyPKIDPairSnapshotEntry, len(uniqueBLSPkidPairSnapshotEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueBLSPkidPairEntries {
		pgBLSPkidPairEntrySlice[ii] = &PGBLSPkidPairEntry{BLSPublicKeyPKIDPairEntry: BLSPublicKeyPKIDPairEncoderToPGStruct(
			entry.Encoder.(*lib.BLSPublicKeyPKIDPairEntry), entry.KeyBytes, params)}
	}

	for ii, entry := range uniqueBLSPkidPairSnapshotEntries {
		pgBLSPkidPairSnapshotEntrySlice[ii] = &PGBLSPublicKeyPKIDPairSnapshotEntry{
			BLSPublicKeyPKIDPairSnapshotEntry: BLSPublicKeyPKIDPairSnapshotEncoderToPGStruct(
				entry.Encoder.(*lib.BLSPublicKeyPKIDPairEntry), entry.KeyBytes, params)}
	}

	if len(pgBLSPkidPairEntrySlice) > 0 {
		// Execute the insert query.
		query := db.NewInsert().Model(&pgBLSPkidPairEntrySlice)

		if operationType == lib.DbOperationTypeUpsert {
			query = query.On("CONFLICT (badger_key) DO UPDATE")
		}

		if _, err := query.Returning("").Exec(context.Background()); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertBLSPkidPairEntry: Error inserting entries")
		}
	}

	if len(pgBLSPkidPairSnapshotEntrySlice) > 0 {
		// Execute query for snapshot entries.
		query := db.NewInsert().Model(&pgBLSPkidPairSnapshotEntrySlice)

		if operationType == lib.DbOperationTypeUpsert {
			query = query.On("CONFLICT (badger_key) DO UPDATE")
		}

		if _, err := query.Returning("").Exec(context.Background()); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertBLSPkidPairEntry: Error inserting snapshot entries")
		}
	}

	// Update the cache with the new entries.
	for _, entry := range uniqueEntries {
		cachedEntries.Add(string(entry.KeyBytes), entry.EncoderBytes)
	}

	return nil
}

// bulkDeleteBLSPkidPairEntry deletes a batch of stake entries from the database.
func bulkDeleteBLSPkidPairEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, cachedEntries *lru.Cache[string, []byte]) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)
	blsPKIDPairEntryKeysToDelete := consumer.FilterKeysByPrefix(keysToDelete,
		lib.Prefixes.PrefixValidatorBLSPublicKeyPKIDPairEntry)
	blsPKIDPairSnapshotEntryKeysToDelete := consumer.FilterKeysByPrefix(keysToDelete,
		lib.Prefixes.PrefixSnapshotValidatorBLSPublicKeyPKIDPairEntry)

	// Execute the delete query.
	if len(blsPKIDPairEntryKeysToDelete) > 0 {
		if _, err := db.NewDelete().
			Model(&PGBLSPkidPairEntry{}).
			Where("badger_key IN (?)", bun.In(blsPKIDPairEntryKeysToDelete)).
			Returning("").
			Exec(context.Background()); err != nil {
			return errors.Wrapf(err, "entries.bulkDeleteBLSPkidPairEntry: Error deleting entries")
		}
	}

	// Execute the delete query for snapshot entries.
	if len(blsPKIDPairSnapshotEntryKeysToDelete) > 0 {
		if _, err := db.NewDelete().
			Model(&PGBLSPublicKeyPKIDPairSnapshotEntry{}).
			Where("badger_key IN (?)", bun.In(blsPKIDPairSnapshotEntryKeysToDelete)).
			Returning("").
			Exec(context.Background()); err != nil {
			return errors.Wrapf(err, "entries.bulkDeleteBLSPkidPairEntry: Error deleting snapshot entries")
		}
	}

	// Remove the deleted entries from the cache.
	for _, key := range keysToDelete {
		cachedEntries.Remove(string(key))
	}

	return nil
}
