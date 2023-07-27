package entries

import (
	"context"
	"encoding/hex"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type DaoCoinLimitOrderEntry struct {
	OrderId                                      string `bun:",nullzero"`
	TransactorPkid                               string `bun:",nullzero"`
	BuyingDaoCoinCreatorPkid                     string `bun:",nullzero"`
	SellingDaoCoinCreatorPkid                    string `bun:",nullzero"`
	ScaledExchangeRateCoinsToSellPerCoinToBuyHex string `bun:",nullzero"`
	QuantityToFillInBaseUnitsHex                 string `bun:",nullzero"`
	OperationType                                uint8  `bun:",nullzero"`
	FillType                                     uint8  `bun:",nullzero"`
	BlockHeight                                  uint32 `bun:",nullzero"`
	BadgerKey                                    []byte `pg:",pk,use_zero"`
}

type PGDaoCoinLimitOrderEntry struct {
	bun.BaseModel `bun:"table:dao_coin_limit_order_entry"`
	DaoCoinLimitOrderEntry
}

// Convert the PostAssociation DeSo encoder to the PG struct used by bun.
func DaoCoinLimitOrderEncoderToPGStruct(daoCoinLimitOrder *lib.DAOCoinLimitOrderEntry, keyBytes []byte) DaoCoinLimitOrderEntry {
	pgEntry := DaoCoinLimitOrderEntry{
		OrderId:                   hex.EncodeToString(daoCoinLimitOrder.OrderID[:]),
		TransactorPkid:            consumer.PublicKeyBytesToBase58Check(daoCoinLimitOrder.TransactorPKID[:]),
		BuyingDaoCoinCreatorPkid:  consumer.PublicKeyBytesToBase58Check(daoCoinLimitOrder.BuyingDAOCoinCreatorPKID[:]),
		SellingDaoCoinCreatorPkid: consumer.PublicKeyBytesToBase58Check(daoCoinLimitOrder.SellingDAOCoinCreatorPKID[:]),
		ScaledExchangeRateCoinsToSellPerCoinToBuyHex: daoCoinLimitOrder.ScaledExchangeRateCoinsToSellPerCoinToBuy.String(),
		QuantityToFillInBaseUnitsHex:                 daoCoinLimitOrder.QuantityToFillInBaseUnits.String(),
		OperationType:                                uint8(daoCoinLimitOrder.OperationType),
		FillType:                                     uint8(daoCoinLimitOrder.FillType),
		BlockHeight:                                  daoCoinLimitOrder.BlockHeight,
		BadgerKey:                                    keyBytes,
	}
	return pgEntry
}

// DaoCoinLimitOrderBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func DaoCoinLimitOrderBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteDaoCoinLimitOrderEntry(entries, db, operationType)
	} else {
		err = bulkInsertDaoCoinLimitOrderEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.DaoCoinLimitOrderBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertDaoCoinLimitOrderEntry inserts a batch of post_association entries into the database.
func bulkInsertDaoCoinLimitOrderEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGDaoCoinLimitOrderEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGDaoCoinLimitOrderEntry{DaoCoinLimitOrderEntry: DaoCoinLimitOrderEncoderToPGStruct(entry.Encoder.(*lib.DAOCoinLimitOrderEntry), entry.KeyBytes)}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertDaoCoinLimitOrderEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of post_association entries from the database.
func bulkDeleteDaoCoinLimitOrderEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGDaoCoinLimitOrderEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteDaoCoinLimitOrderEntry: Error deleting entries")
	}

	return nil
}
