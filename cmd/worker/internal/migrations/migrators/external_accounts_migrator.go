package migrators

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExternalAccountsMigrator is a background job that encrypts
// external accounts data on startup.
// It periodically waits until a keyring is configured to determine
// how many services it must migrate.
// Scheduling and progress report is delegated to the out of band
// migration package.
// The migration is non destructive and can be reverted.
type ExternalAccountsMigrator struct {
	store        *basestore.Store
	BatchSize    int
	AllowDecrypt bool
}

var _ oobmigration.Migrator = &ExternalAccountsMigrator{}

func NewExternalAccountsMigrator(store *basestore.Store) *ExternalAccountsMigrator {
	// not locking too many external accounts at a time to prevent congestion
	return &ExternalAccountsMigrator{store: store, BatchSize: 50}
}

func NewExternalAccountsMigratorWithDB(db dbutil.DB) *ExternalAccountsMigrator {
	return NewExternalAccountsMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID of the migration row in the out_of_band_migrations table.
// This ID was defined arbitrarily in this migration file: frontend/1528395809_external_account_migration.up.sql
func (m *ExternalAccountsMigrator) ID() int {
	return 6
}

// Progress returns a value from 0 to 1 representing the percentage of configuration already migrated.
func (m *ExternalAccountsMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM user_external_accounts WHERE encryption_key_id != '' OR (account_data IS NULL AND auth_data IS NULL)) c1,
			(SELECT COUNT(*) AS count FROM user_external_accounts) c2
	`)))
	return progress, err
}

// Up loads BatchSize external accounts, locks them, and encrypts their config using the
// key returned by keyring.Default().
// If there is no ring, it will periodically try again until the key is setup in the config.
// Up ensures the configuration can be decrypted with the same key before overwitting it.
// The key id is stored alongside the encrypted configuration.
func (m *ExternalAccountsMigrator) Up(ctx context.Context) (err error) {
	key := keyring.Default().UserExternalAccountKey
	if key == nil {
		return nil
	}

	version, err := key.Version(ctx)
	if err != nil {
		return err
	}

	keyIdent := version.JSON()

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalAccountsWith(tx)
	accounts, err := store.ListBySQL(ctx, sqlf.Sprintf("WHERE encryption_key_id = '' AND (account_data IS NOT NULL OR auth_data IS NOT NULL) ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED", m.BatchSize))
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		var (
			encAuthData *string
			encData     *string
		)
		if acc.AuthData != nil {
			encrypted, err := key.Encrypt(ctx, *acc.AuthData)
			if err != nil {
				return err
			}

			// ensure encryption round-trip is valid
			decrypted, err := key.Decrypt(ctx, encrypted)
			if err != nil {
				return err
			}
			if decrypted.Secret() != string(*acc.AuthData) {
				return errors.New("invalid encryption round-trip")
			}

			encAuthData = strptr(string(encrypted))
		}

		if acc.Data != nil {
			encrypted, err := key.Encrypt(ctx, *acc.Data)
			if err != nil {
				return err
			}

			// ensure encryption round-trip is valid
			decrypted, err := key.Decrypt(ctx, encrypted)
			if err != nil {
				return err
			}
			if decrypted.Secret() != string(*acc.Data) {
				return errors.New("invalid encryption round-trip")
			}

			encData = strptr(string(encrypted))
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE user_external_accounts SET auth_data = %s, account_data = %s, encryption_key_id = %s WHERE id = %d",
			encAuthData,
			encData,
			keyIdent,
			acc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func strptr(s string) *string {
	return &s
}

func (m *ExternalAccountsMigrator) Down(ctx context.Context) (err error) {
	key := keyring.Default().UserExternalAccountKey
	if key == nil {
		return nil
	}

	if !m.AllowDecrypt {
		return nil
	}

	// For records that were encrypted, we need to decrypt the configuration,
	// store it in plain text and remove the encryption_key_id.
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	store := database.ExternalAccountsWith(tx)
	accounts, err := store.ListBySQL(ctx, sqlf.Sprintf("WHERE encryption_key_id != '' ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED", m.BatchSize))
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE user_external_accounts SET auth_data = %s, encryption_key_id = '' WHERE id = %s",
			acc.AuthData,
			acc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}
