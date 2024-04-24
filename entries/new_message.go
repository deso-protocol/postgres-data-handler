package entries

import (
	"bytes"
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type NewMessageEntry struct {
	SenderAccessGroupOwnerPublicKey    string `bun:",nullzero"`
	SenderAccessGroupKeyName           string `bun:",nullzero"`
	SenderAccessGroupPublicKey         string `bun:",nullzero"`
	RecipientAccessGroupOwnerPublicKey string `bun:",nullzero"`
	RecipientAccessGroupKeyName        string `bun:",nullzero"`
	RecipientAccessGroupPublicKey      string `bun:",nullzero"`
	EncryptedText                      string `pg:",use_zero"`
	IsGroupChatMessage                 bool
	Timestamp                          time.Time `pg:",use_zero"`

	ExtraData map[string]string `bun:"type:jsonb"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGNewMessageEntry struct {
	bun.BaseModel `bun:"table:new_message_entry"`
	NewMessageEntry
}

type PGNewMessageEntryUtxoOps struct {
	bun.BaseModel `bun:"table:new_message_entry_utxo_ops"`
	NewMessageEntry
	UtxoOperation
}

// Convert the NewMessage DeSo encoder to the PGNewMessageEntry struct used by bun.
func NewMessageEncoderToPGStruct(newMessageEntry *lib.NewMessageEntry, keyBytes []byte, params *lib.DeSoParams) NewMessageEntry {
	isGroupChatMessage := false

	if bytes.Equal(keyBytes[:1], lib.Prefixes.PrefixGroupChatMessagesIndex) {
		isGroupChatMessage = true
	}

	pgNewMessageEntry := NewMessageEntry{
		EncryptedText:      string(newMessageEntry.EncryptedText[:]),
		Timestamp:          consumer.UnixNanoToTime(newMessageEntry.TimestampNanos),
		ExtraData:          consumer.ExtraDataBytesToString(newMessageEntry.ExtraData),
		IsGroupChatMessage: isGroupChatMessage,
		BadgerKey:          keyBytes,
	}

	if newMessageEntry.SenderAccessGroupOwnerPublicKey != nil {
		pgNewMessageEntry.SenderAccessGroupOwnerPublicKey = consumer.PublicKeyBytesToBase58Check(newMessageEntry.SenderAccessGroupOwnerPublicKey[:], params)
	}

	if newMessageEntry.SenderAccessGroupKeyName != nil {
		pgNewMessageEntry.SenderAccessGroupKeyName = string(newMessageEntry.SenderAccessGroupKeyName[:])
	}

	if newMessageEntry.SenderAccessGroupPublicKey != nil {
		pgNewMessageEntry.SenderAccessGroupPublicKey = consumer.PublicKeyBytesToBase58Check(newMessageEntry.SenderAccessGroupPublicKey[:], params)
	}

	if newMessageEntry.RecipientAccessGroupOwnerPublicKey != nil {
		pgNewMessageEntry.RecipientAccessGroupOwnerPublicKey = consumer.PublicKeyBytesToBase58Check(newMessageEntry.RecipientAccessGroupOwnerPublicKey[:], params)
	}

	if newMessageEntry.RecipientAccessGroupKeyName != nil {
		pgNewMessageEntry.RecipientAccessGroupKeyName = string(newMessageEntry.RecipientAccessGroupKeyName[:])
	}

	if newMessageEntry.RecipientAccessGroupPublicKey != nil {
		pgNewMessageEntry.RecipientAccessGroupPublicKey = consumer.PublicKeyBytesToBase58Check(newMessageEntry.RecipientAccessGroupPublicKey[:], params)
	}

	return pgNewMessageEntry
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func NewMessageBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB, params *lib.DeSoParams) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteNewMessageEntry(entries, db, operationType)
	} else {
		err = bulkInsertNewMessageEntry(entries, db, operationType, params)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertNewMessageEntry inserts a batch of new_message entries into the database.
func bulkInsertNewMessageEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType, params *lib.DeSoParams) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGNewMessageEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGNewMessageEntry{NewMessageEntry: NewMessageEncoderToPGStruct(entry.Encoder.(*lib.NewMessageEntry), entry.KeyBytes, params)}
	}

	// Execute the insert query.
	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertNewMessageEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of new_message entries from the database.
func bulkDeleteNewMessageEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGNewMessageEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteNewMessageEntry: Error deleting entries")
	}

	return nil
}
