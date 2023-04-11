package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGProfileEntry struct {
	bun.BaseModel `bun:"table:profile_entry"`
	PublicKey     string `pg:",pk,use_zero" decode_function:"base_58_check" decode_src_field_name:"PublicKey"`
	Pkid          []byte `pg:",use_zero"`
	Username      string `bun:",nullzero" decode_function:"string_bytes" decode_src_field_name:"Username"`
	Description   string `bun:",nullzero" decode_function:"string_bytes" decode_src_field_name:"Description"`
	ProfilePic    []byte `bun:",nullzero"`
	// TODO: Determine if these fields are even necessary, or if the coin entry will be captured elsewhere and can just be joined to this table.
	CreatorBasisPoints               uint64                        `decode_function:"nested_value" decode_src_field_name:"CreatorCoinEntry" nested_field_name:"CreatorBasisPoints"`
	CoinWatermarkNanos               uint64                        `decode_function:"nested_value" decode_src_field_name:"CreatorCoinEntry" nested_field_name:"CoinWatermarkNanos"`
	MintingDisabled                  bool                          `decode_function:"nested_value" decode_src_field_name:"CreatorCoinEntry" nested_field_name:"MintingDisabled"`
	DaoCoinMintingDisabled           bool                          `decode_function:"nested_value" decode_src_field_name:"DAOCoinEntry" nested_field_name:"MintingDisabled"`
	DaoCoinTransferRestrictionStatus lib.TransferRestrictionStatus `decode_function:"nested_value" decode_src_field_name:"DAOCoinEntry" nested_field_name:"TransferRestrictionStatus"`
	ExtraData                        map[string]string             `bun:"type:jsonb" decode_function:"extra_data" decode_src_field_name:"ExtraData"`
	BadgerKey                        []byte                        `pg:",use_zero"`
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func ProfileBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteProfileEntry(entries, db, operationType)
	} else {
		err = bulkInsertProfileEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertPostEntry inserts a batch of post entries into the database.
func bulkInsertProfileEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGProfileEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for i := len(uniqueEntries) - 1; i >= 0; i-- {
		encoder := uniqueEntries[i].Encoder

		pgProfileEntry := &PGProfileEntry{}
		// Copy all encoder fields to the bun struct.
		consumer.CopyStruct(encoder, pgProfileEntry)

		// Add the pkid to the profile entry.
		pkid := consumer.GetPKIDBytesFromKey(uniqueEntries[i].KeyBytes)
		pgProfileEntry.Pkid = pkid

		// Add the badger key to the struct.
		pgProfileEntry.BadgerKey = entries[i].KeyBytes
		pgEntrySlice[i] = pgProfileEntry
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (public_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertProfileEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of profile entries from the database.
func bulkDeleteProfileEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGProfileEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteProfileEntry: Error deleting entries")
	}

	return nil
}
