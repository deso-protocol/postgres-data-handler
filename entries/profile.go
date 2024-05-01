package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type ProfileEntry struct {
	PublicKey                        string `pg:",pk,use_zero"`
	Pkid                             string `pg:",use_zero"`
	Username                         string `bun:",nullzero"`
	Description                      string `bun:",nullzero"`
	ProfilePic                       []byte `bun:",nullzero"`
	CreatorBasisPoints               uint64
	CoinWatermarkNanos               uint64
	MintingDisabled                  bool
	DesoLockedNanos                  uint64
	CcCoinsInCirculationNanos        uint64
	DaoCoinsInCirculationNanosHex    string
	DaoCoinMintingDisabled           bool
	DaoCoinTransferRestrictionStatus lib.TransferRestrictionStatus
	ExtraData                        map[string]string `bun:"type:jsonb"`
	BadgerKey                        []byte            `pg:",use_zero"`
}

type PGProfileEntry struct {
	bun.BaseModel `bun:"table:profile_entry"`
	ProfileEntry
}

type PGProfileEntryUtxoOps struct {
	bun.BaseModel `bun:"table:profile_entry_utxo_ops"`
	ProfileEntry
	UtxoOperation
}

func ProfileEntryEncoderToPGStruct(profileEntry *lib.ProfileEntry, keyBytes []byte, params *lib.DeSoParams) ProfileEntry {
	return ProfileEntry{
		PublicKey:                        consumer.PublicKeyBytesToBase58Check(profileEntry.PublicKey, params),
		Pkid:                             consumer.PublicKeyBytesToBase58Check(consumer.GetPKIDBytesFromKey(keyBytes), params),
		Username:                         string(profileEntry.Username),
		Description:                      string(profileEntry.Description),
		ProfilePic:                       profileEntry.ProfilePic,
		CreatorBasisPoints:               profileEntry.CreatorCoinEntry.CreatorBasisPoints,
		CoinWatermarkNanos:               profileEntry.CreatorCoinEntry.CoinWatermarkNanos,
		MintingDisabled:                  profileEntry.CreatorCoinEntry.MintingDisabled,
		DesoLockedNanos:                  profileEntry.CreatorCoinEntry.DeSoLockedNanos,
		CcCoinsInCirculationNanos:        profileEntry.CreatorCoinEntry.CoinsInCirculationNanos.Uint64(),
		DaoCoinsInCirculationNanosHex:    profileEntry.DAOCoinEntry.CoinsInCirculationNanos.String(),
		DaoCoinMintingDisabled:           profileEntry.DAOCoinEntry.MintingDisabled,
		DaoCoinTransferRestrictionStatus: profileEntry.DAOCoinEntry.TransferRestrictionStatus,
		ExtraData:                        consumer.ExtraDataBytesToString(profileEntry.ExtraData),
		BadgerKey:                        keyBytes,
	}

}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func ProfileBatchOperation(entries []*lib.StateChangeEntry, db bun.IDB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteProfileEntry(entries, db, operationType)
	} else {
		err = bulkInsertProfileEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertPostEntry inserts a batch of post entries into the database.
func bulkInsertProfileEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGProfileEntry, len(uniqueEntries))

	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGProfileEntry{ProfileEntry: ProfileEntryEncoderToPGStruct(entry.Encoder.(*lib.ProfileEntry), entry.KeyBytes, params)}
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
func bulkDeleteProfileEntry(entries []*lib.StateChangeEntry, db bun.IDB, operationType lib.StateSyncerOperationType) error {
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
