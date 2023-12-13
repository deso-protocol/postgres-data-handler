package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
)

// TODO: when to use nullzero vs use_zero?
type ValidatorEntry struct {
	ValidatorPKID                       string   `bun:",nullzero"`
	Domains                             []string `bun:",nullzero"`
	DisableDelegatedStake               bool
	DelegatedStakeCommissionBasisPoints uint64
	VotingPublicKey                     string `bun:",nullzero"`
	VotingAuthorization                 string `bun:",nullzero"`
	// Use bunbig.Int to store the balance as a numeric in the pg database.
	TotalStakeAmountNanos   *bunbig.Int `pg:",use_zero"`
	LastActiveAtEpochNumber uint64
	JailedAtEpochNumber     uint64

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGValidatorEntry struct {
	bun.BaseModel `bun:"table:validator_entry"`
	ValidatorEntry
}

// TODO: Do I need this?
type PGValidatorEntryUtxoOps struct {
	bun.BaseModel `bun:"table:validator_entry_utxo_ops"`
	ValidatorEntry
	UtxoOperation
}

// Convert the ValidatorEntry DeSo encoder to the PGValidatorEntry struct used by bun.
func ValidatorEncoderToPGStruct(validatorEntry *lib.ValidatorEntry, keyBytes []byte, params *lib.DeSoParams) ValidatorEntry {
	pgValidatorEntry := ValidatorEntry{
		ExtraData: consumer.ExtraDataBytesToString(validatorEntry.ExtraData),
		BadgerKey: keyBytes,
	}

	if validatorEntry.ValidatorPKID != nil {
		pgValidatorEntry.ValidatorPKID = consumer.PublicKeyBytesToBase58Check((*validatorEntry.ValidatorPKID)[:], params)
	}

	if validatorEntry.Domains != nil {
		pgValidatorEntry.Domains = make([]string, len(validatorEntry.Domains))
		for ii, domain := range validatorEntry.Domains {
			pgValidatorEntry.Domains[ii] = string(domain)
		}
	}

	pgValidatorEntry.DisableDelegatedStake = validatorEntry.DisableDelegatedStake
	pgValidatorEntry.DelegatedStakeCommissionBasisPoints = validatorEntry.DelegatedStakeCommissionBasisPoints

	if validatorEntry.VotingPublicKey != nil {
		pgValidatorEntry.VotingPublicKey = validatorEntry.VotingPublicKey.ToString()
	}

	if validatorEntry.VotingAuthorization != nil {
		pgValidatorEntry.VotingAuthorization = validatorEntry.VotingAuthorization.ToString()
	}

	pgValidatorEntry.TotalStakeAmountNanos = bunbig.FromMathBig(validatorEntry.TotalStakeAmountNanos.ToBig())
	pgValidatorEntry.LastActiveAtEpochNumber = validatorEntry.LastActiveAtEpochNumber
	pgValidatorEntry.JailedAtEpochNumber = validatorEntry.JailedAtEpochNumber

	return pgValidatorEntry
}

// ValidatorBatchOperation is the entry point for processing a batch of Validator entries.
// It determines the appropriate handler based on the operation type and executes it.
func ValidatorBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteValidatorEntry(entries, db, operationType)
	} else {
		err = bulkInsertValidatorEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.ValidatorBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertValidatorEntry inserts a batch of validator entries into the database.
func bulkInsertValidatorEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGValidatorEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGValidatorEntry{ValidatorEntry: ValidatorEncoderToPGStruct(entry.Encoder.(*lib.ValidatorEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertValidatorEntry: Error inserting entries")
	}
	return nil
}

// bulkDeleteValidatorEntry deletes a batch of validator entries from the database.
func bulkDeleteValidatorEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGValidatorEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteValidatorEntry: Error deleting entries")
	}

	return nil
}
