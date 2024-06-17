package entries

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// GetDbHandle returns the correct interface to use for database operations.
// If a transaction is open, it returns the transaction handle, otherwise it returns the db handle.
func GetDbHandle(tx *bun.Tx, db *bun.DB) bun.IDB {
	if tx != nil {
		return tx
	}
	return db
}

// CreateSavepoint creates a savepoint in the current transaction. If no transaction is open, it returns an empty string.
// The randomly generated savepoint name is returned if the savepoint is created successfully.
func CreateSavepoint(tx *bun.Tx) (string, error) {
	if tx == nil {
		return "", nil
	}
	savepointName := uuid.New().String()

	_, err := tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepointName))
	if err != nil {
		return "", errors.Wrapf(err, "PostgresDataHandler.CreateSavepoint: Error creating savepoint")
	}

	return savepointName, nil
}

func RollbackToSavepoint(tx *bun.Tx, savepointName string) error {
	if tx == nil || savepointName == "" {
		return nil
	}
	_, err := tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName))
	if err != nil {
		return errors.Wrapf(err, "PostgresDataHandler.RollbackToSavepoint: Error reverting to savepoint")
	}
	return nil
}
