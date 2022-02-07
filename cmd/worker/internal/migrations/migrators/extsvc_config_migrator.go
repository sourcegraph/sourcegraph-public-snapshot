package migrators

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExternalServiceConfigMigrator is a background job that encrypts
// external services config on startup.
// It periodically waits until a keyring is configured to determine
// how many services it must migrate.
// Scheduling and progress report is deleguated to the out of band
// migration package.
// The migration is non destructive and can be reverted.
type ExternalServiceConfigMigrator struct {
	store        *basestore.Store
	BatchSize    int
	AllowDecrypt bool
}

var _ oobmigration.Migrator = &ExternalServiceConfigMigrator{}

func NewExternalServiceConfigMigrator(store *basestore.Store) *ExternalServiceConfigMigrator {
	// not locking too many external services at a time to prevent congestion
	return &ExternalServiceConfigMigrator{store: store, BatchSize: 50}
}

func NewExternalServiceConfigMigratorWithDB(db dbutil.DB) *ExternalServiceConfigMigrator {
	return NewExternalServiceConfigMigrator(basestore.NewWithDB(db, sql.TxOptions{}))
}

// ID of the migration row in in the out_of_band_migrations table.
// This ID was defined arbitrarily in this migration file: frontend/1528395802_external_service_config_migration.up.sql.
func (m *ExternalServiceConfigMigrator) ID() int {
	return 3
}

// Progress returns a value from 0 to 1 representing the percentage of configuration already migrated.
func (m *ExternalServiceConfigMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE
				CAST(c1.count AS float) / CAST(c2.count AS float)
			END
		FROM
			(SELECT COUNT(*) AS count FROM external_services WHERE encryption_key_id != '') c1,
			(SELECT COUNT(*) AS count FROM external_services) c2
	`)))
	return progress, err
}

// Up loads BatchSize external services, locks them, and encrypts their config using the
// key returned by keyring.Default().
// If there is no ring, it will periodically try again until the key is setup in the config.
// Up ensures the configuration can be decrypted with the same key before overwitting it.
// The key id is stored alongside the encrypted configuration.
func (m *ExternalServiceConfigMigrator) Up(ctx context.Context) (err error) {
	key := keyring.Default().ExternalServiceKey
	if key == nil {
		return nil
	}

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	services, err := m.listConfigsForUpdate(ctx, tx, false)
	if err != nil {
		return err
	}

	for _, svc := range services {
		encryptedCfg, err := key.Encrypt(ctx, []byte(svc.Config))
		if err != nil {
			return err
		}

		version, err := key.Version(ctx)
		if err != nil {
			return err
		}
		keyIdent := version.JSON()

		// ensure encryption round-trip is valid with keyIdent
		decrypted, err := key.Decrypt(ctx, encryptedCfg)
		if err != nil {
			return err
		}
		if decrypted.Secret() != svc.Config {
			return errors.New("invalid encryption round-trip")
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE external_services SET config = %s, encryption_key_id = %s WHERE id = %s",
			encryptedCfg,
			keyIdent,
			svc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (m *ExternalServiceConfigMigrator) Down(ctx context.Context) (err error) {
	key := keyring.Default().ExternalServiceKey
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

	services, err := m.listConfigsForUpdate(ctx, tx, true)
	if err != nil {
		return err
	}

	for _, svc := range services {
		secret, err := key.Decrypt(ctx, []byte(svc.Config))
		if err != nil {
			return err
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE external_services SET config = %s, encryption_key_id = '' WHERE id = %s",
			secret.Secret(),
			svc.ID,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (m *ExternalServiceConfigMigrator) listConfigsForUpdate(ctx context.Context, tx *basestore.Store, encrypted bool) ([]*types.ExternalService, error) {
	// Select and lock a few records within this transaction. This ensures
	// that many frontend instances can run the same migration concurrently
	// without them all trying to convert the same record.
	q := "SELECT id, config FROM external_services "
	if encrypted {
		q += "WHERE encryption_key_id != ''"
	} else {
		q += "WHERE encryption_key_id = ''"
	}

	q += "ORDER BY id ASC LIMIT %s FOR UPDATE SKIP LOCKED"

	rows, err := tx.Query(ctx, sqlf.Sprintf(q, m.BatchSize))

	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var services []*types.ExternalService

	for rows.Next() {
		var svc types.ExternalService
		if err := rows.Scan(&svc.ID, &svc.Config); err != nil {
			return nil, err
		}
		services = append(services, &svc)
	}

	return services, nil
}
