package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PkidEntry struct {
	Pkid      string `pg:",use_zero"`
	PublicKey string `pg:",use_zero"`
	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGPkidEntry struct {
	bun.BaseModel `bun:"table:pkid_entry"`
	PkidEntry
}

type PGPkidEntryUtxoOps struct {
	bun.BaseModel `bun:"table:pkid_entry_utxo_ops"`
	PkidEntry
	UtxoOperation
}

type LeaderScheduleEntry struct {
	SnapshotAtEpochNumber uint64 `pg:",use_zero"`
	LeaderIndex           uint16 `pg:",use_zero"`
	ValidatorPKID         string `pg:",use_zero"`
	BadgerKey             []byte `pg:",pk,use_zero"`
}

type PGLeaderScheduleEntry struct {
	bun.BaseModel `bun:"table:leader_schedule_entry"`
	LeaderScheduleEntry
}

// Convert the Diamond DeSo encoder to the PG struct used by bun.
func PkidEntryEncoderToPGStruct(pkidEntry *lib.PKIDEntry, keyBytes []byte, params *lib.DeSoParams) PkidEntry {
	return PkidEntry{
		Pkid:      consumer.PublicKeyBytesToBase58Check(pkidEntry.PKID[:], params),
		PublicKey: consumer.PublicKeyBytesToBase58Check(pkidEntry.PublicKey[:], params),
		BadgerKey: keyBytes,
	}
}

// Convert the leader schedule entry to the PG struct used by bun.
func LeaderScheduleEncoderToPGStruct(validatorPKID *lib.PKID, keyBytes []byte, params *lib.DeSoParams,
) *LeaderScheduleEntry {
	prefixRemovedKeyBytes := keyBytes[1:]
	if len(prefixRemovedKeyBytes) != 10 {
		glog.Errorf("LeaderScheduleEncoderToPGStruct: Invalid key length: %d", len(prefixRemovedKeyBytes))
		return nil
	}
	epochNumber := lib.DecodeUint64(prefixRemovedKeyBytes[:8])
	leaderIndex := lib.DecodeUint16(prefixRemovedKeyBytes[8:10])
	return &LeaderScheduleEntry{
		ValidatorPKID:         consumer.PublicKeyBytesToBase58Check(validatorPKID[:], params),
		SnapshotAtEpochNumber: epochNumber,
		LeaderIndex:           leaderIndex,
		BadgerKey:             keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func PkidEntryBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeletePkidEntry(entries, db, operationType)
	} else {
		err = bulkInsertPkidEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertDiamondEntry inserts a batch of diamond entries into the database.
func bulkInsertPkidEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGPkidEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGPkidEntry{PkidEntry: PkidEntryEncoderToPGStruct(entry.Encoder.(*lib.PKIDEntry), entry.KeyBytes, params)}
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

func PkidBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeletePkid(entries, db, operationType)
	} else {
		err = bulkInsertPkid(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertPkid inserts a batch of PKIDs into the database.
func bulkInsertPkid(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	uniqueLeaderScheduleEntries := consumer.FilterEntriesByPrefix(
		uniqueEntries, lib.Prefixes.PrefixSnapshotLeaderSchedule)
	// NOTE: if we need to support parsing other indexes for PKIDs beyond LeaderSchedule,
	// we will need to filter the uniqueEntries by the appropriate prefix and then convert
	// the entries to the appropriate PG struct.
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLeaderScheduleEntry, len(uniqueLeaderScheduleEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueLeaderScheduleEntries {
		leaderScheduleEntry := LeaderScheduleEncoderToPGStruct(entry.Encoder.(*lib.PKID), entry.KeyBytes, params)
		if leaderScheduleEntry == nil {
			glog.Errorf("bulkInsertPkid: Error converting LeaderScheduleEntry to PG struct")
			continue
		}
		pgEntrySlice[ii] = &PGLeaderScheduleEntry{LeaderScheduleEntry: *leaderScheduleEntry}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertPkid: Error inserting entries")
	}
	return nil
}

// bulkDeletePKID deletes a batch of PKIDs from the database.
func bulkDeletePkid(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)
	leaderSchedKeysToDelete := consumer.FilterKeysByPrefix(keysToDelete, lib.Prefixes.PrefixSnapshotLeaderSchedule)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLeaderScheduleEntry{}).
		Where("badger_key IN (?)", bun.In(leaderSchedKeysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeletePkid: Error deleting entries")
	}

	return nil
}
