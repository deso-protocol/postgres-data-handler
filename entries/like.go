package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type LikeEntry struct {
	PublicKey string `pg:",use_zero"`
	PostHash  string `pg:",use_zero"`
	BadgerKey []byte `pg:",pk,use_zero"`
}

type PGLikeEntry struct {
	bun.BaseModel `bun:"table:like_entry"`
	LikeEntry
}

type PGLikeEntryUtxoOps struct {
	bun.BaseModel `bun:"table:like_entry_utxo_ops"`
	LikeEntry
	UtxoOperation
}

// Convert the like DeSo encoder to the PG struct used by bun.
func LikeEncoderToPGStruct(likeEntry *lib.LikeEntry, keyBytes []byte, params *lib.DeSoParams) LikeEntry {
	return LikeEntry{
		PublicKey: consumer.PublicKeyBytesToBase58Check(likeEntry.LikerPubKey[:], params),
		PostHash:  hex.EncodeToString(likeEntry.LikedPostHash[:]),
		BadgerKey: keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func LikeBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteLikeEntry(entries, db, operationType)
	} else {
		err = bulkInsertLikeEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertLikeEntry inserts a batch of like entries into the database.
func bulkInsertLikeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGLikeEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGLikeEntry{LikeEntry: LikeEncoderToPGStruct(entry.Encoder.(*lib.LikeEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertLikeEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of like entries from the database.
func bulkDeleteLikeEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGLikeEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteLikeEntry: Error deleting entries")
	}

	return nil
}
