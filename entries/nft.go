package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type NftEntry struct {
	LastOwnerPkid              string `pg:",use_zero"`
	OwnerPkid                  string `pg:",use_zero"`
	NftPostHash                string `pg:",use_zero"`
	SerialNumber               uint64 `pg:",use_zero"`
	IsForSale                  bool   `pg:",use_zero"`
	MinBidAmountNanos          uint64 `bun:",nullzero"`
	UnlockableText             string `bun:",nullzero"`
	LastAcceptedBidAmountNanos uint64 `bun:",nullzero"`
	IsPending                  bool   `pg:",use_zero"`
	IsBuyNow                   bool   `pg:",use_zero"`
	BuyNowPriceNanos           uint64 `bun:",nullzero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGNftEntry struct {
	bun.BaseModel `bun:"table:nft_entry"`
	NftEntry
}
type PGNftEntryUtxoOps struct {
	bun.BaseModel `bun:"table:nft_entry_utxo_ops"`
	NftEntry
	UtxoOperation
}

// Convert the NFT DeSo entry into a bun struct.
func NftEncoderToPGStruct(nftEntry *lib.NFTEntry, keyBytes []byte) NftEntry {
	pgNFTEntry := NftEntry{
		OwnerPkid:                  consumer.PublicKeyBytesToBase58Check(nftEntry.OwnerPKID[:]),
		NftPostHash:                hex.EncodeToString(nftEntry.NFTPostHash[:]),
		SerialNumber:               nftEntry.SerialNumber,
		IsForSale:                  nftEntry.IsForSale,
		MinBidAmountNanos:          nftEntry.MinBidAmountNanos,
		LastAcceptedBidAmountNanos: nftEntry.LastAcceptedBidAmountNanos,
		IsPending:                  nftEntry.IsPending,
		IsBuyNow:                   nftEntry.IsBuyNow,
		BuyNowPriceNanos:           nftEntry.BuyNowPriceNanos,
		ExtraData:                  consumer.ExtraDataBytesToString(nftEntry.ExtraData),
		BadgerKey:                  keyBytes,
	}
	if nftEntry.LastOwnerPKID != nil {
		pgNFTEntry.LastOwnerPkid = consumer.PublicKeyBytesToBase58Check(nftEntry.LastOwnerPKID[:])
	}
	if nftEntry.UnlockableText != nil {
		pgNFTEntry.UnlockableText = string(nftEntry.UnlockableText)
	}
	return pgNFTEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func NftBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteNftEntry(entries, db, operationType)
	} else {
		err = bulkInsertNftEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertNftEntry inserts a batch of nft entries into the database.
func bulkInsertNftEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGNftEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGNftEntry{NftEntry: NftEncoderToPGStruct(entry.Encoder.(*lib.NFTEntry), entry.KeyBytes)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertNftEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of nft entries from the database.
func bulkDeleteNftEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGNftEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteNftEntry: Error deleting entries")
	}

	return nil
}
