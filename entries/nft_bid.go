package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type PGNftBidEntry struct {
	bun.BaseModel       `bun:"table:nft_bid_entry"`
	BidderPkid          []byte `pg:",use_zero" decode_function:"pkid" decode_src_field_name:"BidderPKID"`
	NftPostHash         string `pg:",use_zero" decode_function:"blockhash" decode_src_field_name:"NFTPostHash"`
	SerialNumber        uint64 `pg:",use_zero"`
	BidAmountNanos      uint64 `pg:",use_zero"`
	AcceptedBlockHeight uint64 `bun:",nullzero"`
	BadgerKey           []byte `pg:",pk,use_zero"`
}

// Convert the NFT DeSo entry into a bun struct.
func NftBidEncoderToPGStruct(nftBidEntry *lib.NFTBidEntry, keyBytes []byte) *PGNftBidEntry {
	pgNftEntry := &PGNftBidEntry{
		BidderPkid:     nftBidEntry.BidderPKID[:],
		NftPostHash:    hex.EncodeToString(nftBidEntry.NFTPostHash[:]),
		SerialNumber:   nftBidEntry.SerialNumber,
		BidAmountNanos: nftBidEntry.BidAmountNanos,
		BadgerKey:      keyBytes,
	}

	if nftBidEntry.AcceptedBlockHeight != nil {
		pgNftEntry.AcceptedBlockHeight = uint64(*nftBidEntry.AcceptedBlockHeight)
	}
	return pgNftEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func NftBidBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteNftBidEntry(entries, db, operationType)
	} else {
		err = bulkInsertNftBidEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertNftBidEntry inserts a batch of nft_bid entries into the database.
func bulkInsertNftBidEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGNftBidEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = NftBidEncoderToPGStruct(entry.Encoder.(*lib.NFTBidEntry), entry.KeyBytes)
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertNftBidEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of nft_bid entries from the database.
func bulkDeleteNftBidEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGNftBidEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteNftBidEntry: Error deleting entries")
	}

	return nil
}
