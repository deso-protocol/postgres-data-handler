package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type PGPostEntry struct {
	bun.BaseModel                               `bun:"table:post_entry"`
	PostHash                                    string            `pg:",pk,use_zero" decode_function:"blockhash" decode_src_field_name:"PostHash"`
	PosterPublicKey                             string            `pg:",use_zero" decode_function:"base_58_check" decode_src_field_name:"PosterPublicKey"`
	ParentPostHash                              string            `bun:",nullzero" decode_function:"bytehash" decode_src_field_name:"ParentStakeID"`
	Body                                        string            `bun:",nullzero" decode_function:"deso_body_schema" decode_src_field_name:"Body" decode_body_field_name:"Body" decode_image_urls_field_name:"ImageUrls" decode_video_urls_field_name:"VideoUrls"`
	ImageUrls                                   []string          `pg:",nullzero" bun:"type:varchar(255)[]"`
	VideoUrls                                   []string          `pg:",nullzero" bun:"type:varchar(255)[]"`
	RepostedPostHash                            string            `bun:",nullzero" decode_function:"blockhash" decode_src_field_name:"RepostedPostHash"`
	QuotedRepost                                bool              `pg:",use_zero"`
	Timestamp                                   time.Time         `pg:",use_zero" decode_function:"timestamp" decode_src_field_name:"TimestampNanos"`
	Hidden                                      bool              `pg:",use_zero"`
	LikeCount                                   uint64            `pg:",use_zero"`
	RepostCount                                 uint64            `pg:",use_zero"`
	QuoteRepostCount                            uint64            `pg:",use_zero"`
	DiamondCount                                uint64            `pg:",use_zero"`
	CommentCount                                uint64            `pg:",use_zero"`
	Pinned                                      bool              `pg:",use_zero"`
	IsNFT                                       bool              `pg:",use_zero"`
	NumNFTCopies                                uint64            `pg:",use_zero"`
	NumNFTCopiesForSale                         uint64            `pg:",use_zero"`
	NumNFTCopiesBurned                          uint64            `pg:",use_zero"`
	HasUnlockable                               bool              `pg:",use_zero"`
	CreatorRoyaltyBasisPoints                   uint64            `pg:",use_zero"`
	CoinRoyaltyBasisPoints                      uint64            `pg:",use_zero"`
	AdditionalNFTRoyaltiesToCoinsBasisPoints    map[string]uint64 `pg:"additional_nft_royalties_to_coins_basis_points,use_zero" bun:"type:jsonb"`
	AdditionalNFTRoyaltiesToCreatorsBasisPoints map[string]uint64 `pg:"additional_nft_royalties_to_creators_basis_points,use_zero" bun:"type:jsonb"`
	ExtraData                                   map[string]string `bun:"type:jsonb" decode_function:"extra_data" decode_src_field_name:"PostExtraData"`
	IsFrozen                                    bool              `pg:",use_zero"`
	BadgerKey                                   []byte            `pg:",use_zero"`
}

func PostBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = BulkDeletePostEntry(entries, db, operationType)
	} else {
		err = BulkInsertPostEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// TODO: For inserts, have this one run in a non-blocking thread, allow parallelism.
func BulkInsertPostEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGPostEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for i := len(uniqueEntries) - 1; i >= 0; i-- {
		encoder := uniqueEntries[i].Encoder
		pgPostEntry := &PGPostEntry{}
		// Copy all encoder fields to the bun struct.
		consumer.CopyStruct(encoder, pgPostEntry)
		// Add the badger key to the struct.
		pgPostEntry.BadgerKey = entries[i].KeyBytes
		pgEntrySlice[i] = pgPostEntry
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (post_hash) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.BulkInsertPostEntry: Error inserting entries")
	}
	return nil
}

func BulkDeletePostEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	if _, err := db.NewDelete().
		Model(&PGPostEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.BulkDeletePostEntry: Error deleting entries")
	}

	return nil
}
