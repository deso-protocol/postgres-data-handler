package entries

import (
	"context"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"time"
)

type MessageEntry struct {
	SenderPublicKey             string    `pg:",use_zero"`
	RecipientPublicKey          string    `pg:",use_zero"`
	EncryptedText               string    `pg:",use_zero"`
	Timestamp                   time.Time `pg:",use_zero"`
	Version                     uint8     `pg:",use_zero"`
	SenderMessagingPublicKey    string    `pg:",use_zero"`
	RecipientMessagingPublicKey string    `pg:",use_zero"`

	SenderMessagingGroupKeyName    string `pg:",use_zero"`
	RecipientMessagingGroupKeyName string `pg:",use_zero"`

	ExtraData map[string]string `bun:"type:jsonb" decode_function:"extra_data" decode_src_field_name:"ExtraData"`
	BadgerKey []byte            `pg:",pk,use_zero"`
}

type PGMessageEntry struct {
	bun.BaseModel `bun:"table:message_entry"`
	MessageEntry
}

type PGMessageEntryUtxoOps struct {
	bun.BaseModel `bun:"table:message_entry_utxo_ops"`
	MessageEntry
	UtxoOperation
}

// Convert the Message DeSo encoder to the PGMessageEntry struct used by bun.
func MessageEncoderToPGStruct(messageEntry *lib.MessageEntry, keyBytes []byte) MessageEntry {
	return MessageEntry{
		SenderPublicKey:                consumer.PublicKeyBytesToBase58Check(messageEntry.SenderPublicKey[:]),
		RecipientPublicKey:             consumer.PublicKeyBytesToBase58Check(messageEntry.RecipientPublicKey[:]),
		EncryptedText:                  string(messageEntry.EncryptedText),
		Timestamp:                      consumer.UnixNanoToTime(messageEntry.TstampNanos),
		Version:                        messageEntry.Version,
		SenderMessagingPublicKey:       consumer.PublicKeyBytesToBase58Check(messageEntry.SenderMessagingPublicKey[:]),
		RecipientMessagingPublicKey:    consumer.PublicKeyBytesToBase58Check(messageEntry.RecipientMessagingPublicKey[:]),
		SenderMessagingGroupKeyName:    string(messageEntry.SenderMessagingGroupKeyName[:]),
		RecipientMessagingGroupKeyName: string(messageEntry.RecipientMessagingGroupKeyName[:]),
		ExtraData:                      consumer.ExtraDataBytesToString(messageEntry.ExtraData),
		BadgerKey:                      keyBytes,
	}
}

// PostBatchOperation is the entry point for processing a batch of post entries. It determines the appropriate handler
// based on the operation type and executes it.
func MessageBatchOperation(entries []*lib.StateChangeEntry, db *bun.DB) error {
	// We check before we call this function that there is at least one operation type.
	// We also ensure before this that all entries have the same operation type.
	operationType := entries[0].OperationType
	var err error
	if operationType == lib.DbOperationTypeDelete {
		err = bulkDeleteMessageEntry(entries, db, operationType)
	} else {
		err = bulkInsertMessageEntry(entries, db, operationType)
	}
	if err != nil {
		return errors.Wrapf(err, "entries.PostBatchOperation: Problem with operation type %v", operationType)
	}
	return nil
}

// bulkInsertMessageEntry inserts a batch of message entries into the database.
func bulkInsertMessageEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)
	// Create a new array to hold the bun struct.
	pgEntrySlice := make([]*PGMessageEntry, len(uniqueEntries))

	// Loop through the entries and convert them to PGPostEntry.
	for ii, entry := range uniqueEntries {
		pgEntrySlice[ii] = &PGMessageEntry{MessageEntry: MessageEncoderToPGStruct(entry.Encoder.(*lib.MessageEntry), entry.KeyBytes)}
	}

	query := db.NewInsert().Model(&pgEntrySlice)

	if operationType == lib.DbOperationTypeUpsert {
		query = query.On("CONFLICT (badger_key) DO UPDATE")
	}

	if _, err := query.Returning("").Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkInsertMessageEntry: Error inserting entries")
	}
	return nil
}

// bulkDeletePostEntry deletes a batch of message entries from the database.
func bulkDeleteMessageEntry(entries []*lib.StateChangeEntry, db *bun.DB, operationType lib.StateSyncerOperationType) error {
	// Track the unique entries we've inserted so we don't insert the same entry twice.
	uniqueEntries := consumer.UniqueEntries(entries)

	// Transform the entries into a list of keys to delete.
	keysToDelete := consumer.KeysToDelete(uniqueEntries)

	// Execute the delete query.
	if _, err := db.NewDelete().
		Model(&PGMessageEntry{}).
		Where("badger_key IN (?)", bun.In(keysToDelete)).
		Returning("").
		Exec(context.Background()); err != nil {
		return errors.Wrapf(err, "entries.bulkDeleteMessageEntry: Error deleting entries")
	}

	return nil
}
