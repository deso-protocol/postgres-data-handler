package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type PostEntry struct {
	PostHash                                    string            `pg:",pk,use_zero"`
	PosterPublicKey                             string            `pg:",use_zero"`
	ParentPostHash                              string            `bun:",nullzero"`
	Body                                        string            `bun:",nullzero"`
	ImageUrls                                   []string          `pg:",nullzero" bun:"type:varchar(255)[]"`
	VideoUrls                                   []string          `pg:",nullzero" bun:"type:varchar(255)[]"`
	RepostedPostHash                            string            `bun:",nullzero"`
	IsQuotedRepost                              bool              `pg:",use_zero"`
	Timestamp                                   time.Time         `pg:",use_zero"`
	IsHidden                                    bool              `pg:",use_zero"`
	IsPinned                                    bool              `pg:",use_zero"`
	IsNFT                                       bool              `pg:",use_zero"`
	NumNFTCopies                                uint64            `pg:",use_zero"`
	NumNFTCopiesForSale                         uint64            `pg:",use_zero"`
	NumNFTCopiesBurned                          uint64            `pg:",use_zero"`
	HasUnlockable                               bool              `pg:",use_zero"`
	NFTRoyaltyToCreatorBasisPoints              uint64            `pg:",use_zero"`
	NFTRoyaltyToCoinBasisPoints                 uint64            `pg:",use_zero"`
	AdditionalNFTRoyaltiesToCreatorsBasisPoints map[string]uint64 `pg:"additional_nft_royalties_to_creators_basis_points,use_zero" bun:"type:jsonb"`
	AdditionalNFTRoyaltiesToCoinsBasisPoints    map[string]uint64 `pg:"additional_nft_royalties_to_coins_basis_points,use_zero" bun:"type:jsonb"`
	ExtraData                                   map[string]string `bun:"type:jsonb"`
	IsFrozen                                    bool              `pg:",use_zero"`
	BadgerKey                                   []byte            `pg:",use_zero"`
}

type PGPostEntry struct {
	bun.BaseModel `bun:"table:post_entry"`
	PostEntry
}

type PGPostEntryUtxoOps struct {
	bun.BaseModel `bun:"table:post_entry_utxo_ops"`
	PostEntry
	UtxoOperation
}

func PostEntryEncoderToPGStruct(postEntry *lib.PostEntry, keyBytes []byte, params *lib.DeSoParams) (PostEntry, error) {

	pgPostEntry := PostEntry{
		PostHash:                                 hex.EncodeToString(postEntry.PostHash[:]),
		PosterPublicKey:                          consumer.PublicKeyBytesToBase58Check(postEntry.PosterPublicKey, params),
		ParentPostHash:                           hex.EncodeToString(postEntry.ParentStakeID),
		IsQuotedRepost:                           postEntry.IsQuotedRepost,
		Timestamp:                                consumer.UnixNanoToTime(postEntry.TimestampNanos),
		IsHidden:                                 postEntry.IsHidden,
		IsPinned:                                 postEntry.IsPinned,
		IsNFT:                                    postEntry.IsNFT,
		NumNFTCopies:                             postEntry.NumNFTCopies,
		NumNFTCopiesForSale:                      postEntry.NumNFTCopiesForSale,
		NumNFTCopiesBurned:						  postEntry.NumNFTCopiesBurned,
		HasUnlockable:                            postEntry.HasUnlockable,
		NFTRoyaltyToCreatorBasisPoints:           postEntry.NFTRoyaltyToCreatorBasisPoints,
		NFTRoyaltyToCoinBasisPoints:              postEntry.NFTRoyaltyToCoinBasisPoints,
		AdditionalNFTRoyaltiesToCoinsBasisPoints: consumer.ConvertRoyaltyMapToByteStrings(postEntry.AdditionalNFTRoyaltiesToCoinsBasisPoints),
		AdditionalNFTRoyaltiesToCreatorsBasisPoints: consumer.ConvertRoyaltyMapToByteStrings(postEntry.AdditionalNFTRoyaltiesToCreatorsBasisPoints),
		ExtraData: consumer.ExtraDataBytesToString(postEntry.PostExtraData),
		IsFrozen:  postEntry.IsFrozen,
		BadgerKey: keyBytes,
	}

	if postEntry.RepostedPostHash != nil {
		pgPostEntry.RepostedPostHash = hex.EncodeToString(postEntry.RepostedPostHash[:])
	}

	if postEntry.Body != nil {
		// Decode body and image/video urls.
		postBody, err := consumer.DecodeDesoBodySchema(postEntry.Body)
		if err == nil {
			pgPostEntry.Body = postBody.Body
			pgPostEntry.ImageUrls = postBody.ImageURLs
			pgPostEntry.VideoUrls = postBody.VideoURLs
		}
	}

	return pgPostEntry, nil
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func PostBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeletePostEntry(entries, db, operationType)
	} else {
		err = bulkInsertPostEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertPostEntry inserts a batch of post entries into the database.
func bulkInsertPostEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGPostEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		if pgEntry, err := PostEntryEncoderToPGStruct(entry.Encoder.(*lib.PostEntry), entry.KeyBytes, params); err != nil {
			return errors.Wrapf(err, "entries.bulkInsertPostEntry: Problem converting post entry to PG struct")
		} else {
			pgEntrySlice[ii] = &PGPostEntry{PostEntry: pgEntry}
		}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (post_hash) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertPostEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of post entries from the database.
func bulkDeletePostEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGPostEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeletePostEntry: Error deleting entries")
	}

	return nil
}
