package initial_migrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
)

// TODO: Not nullable fields
func createValidatorEntryTable(db *bun.DB, tableName string) error {
	_, err := db.Exec(strings.Replace(`
			CREATE TABLE {tableName} (
				validator_pkid VARCHAR NOT NULL,
				domains VARCHAR ARRAY,
				disable_delegated_stake BOOLEAN,
				delegated_stake_commission_basis_points BIGINT,
				voting_public_key VARCHAR,
				voting_authorization VARCHAR,
				total_stake_amount_nanos NUMERIC(78, 0) NOT NULL,
				last_active_at_epoch_number BIGINT,
				jailed_at_epoch_number BIGINT,
				extra_data JSONB,
				badger_key BYTEA PRIMARY KEY 
			);
			CREATE INDEX {tableName}_validator_pkid_idx ON {tableName} (validator_pkid);
		`, "{tableName}", tableName, -1))
	// TODO: What other fields do we need indexed?
	return err
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		return createValidatorEntryTable(db, "validator_entry")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS validator_entry;
		`)
		if err != nil {
			return err
		}
		return nil
	})
}
