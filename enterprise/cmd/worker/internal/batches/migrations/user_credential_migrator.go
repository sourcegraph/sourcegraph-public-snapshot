package migrations

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const userCredentialMigrationCountPerRun = 5

type userCredentialMigrator struct {
	store        *store.Store
	allowDecrypt bool
}

var _ oobmigration.Migrator = &userCredentialMigrator{}

func (m *userCredentialMigrator) Progress(ctx context.Context) (float64, error) {
	// What is progress, anyway?
	//
	// In this case, there are two things we're trying to do in this migrator.
	// If encryption is enabled, then we want to pick up and encrypt any
	// unencrypted credentials. We also want to replace any
	// "previously-migrated" encryption key IDs with real key ID values.

	progress, _, err := basestore.ScanFirstFloat(
		m.store.Query(ctx, sqlf.Sprintf(
			userCredentialMigratorProgressQuery,
			database.UserCredentialDomainBatches,
			database.UserCredentialPlaceholderEncryptionKeyID,
			database.UserCredentialUnmigratedEncryptionKeyID,
			database.UserCredentialDomainBatches,
		)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const userCredentialMigratorProgressQuery = `
-- source: enterprise/internal/batches/user_credential_migrator.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE CAST((c2.count - c1.count) AS float) / CAST(c2.count AS float) END FROM
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s AND encryption_key_id IN (%s, %s)) c1,
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s) c2
`

func (m *userCredentialMigrator) Up(ctx context.Context) error {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}

	f := func() error {
		credentials, _, err := tx.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope: database.UserCredentialScope{
				Domain: database.UserCredentialDomainBatches,
			},
			LimitOffset: &database.LimitOffset{
				Limit: userCredentialMigrationCountPerRun,
			},
			ForUpdate:         true,
			RequiresMigration: true,
		})
		if err != nil {
			return errors.Wrap(err, "listing user credentials")
		}
		for _, cred := range credentials {
			a, err := cred.Authenticator(ctx)
			if err != nil {
				return errors.Wrapf(err, "retrieving authenticator for ID %d", cred.ID)
			}

			if err := cred.SetAuthenticator(ctx, a); err != nil {
				return errors.Wrapf(err, "setting authenticator for ID %d", cred.ID)
			}

			if err := tx.UserCredentials().Update(ctx, cred); err != nil {
				return errors.Wrapf(err, "upserting user credential %d", cred.ID)
			}
		}

		return nil
	}
	return tx.Done(f())
}

func (m *userCredentialMigrator) Down(ctx context.Context) error {
	if !m.allowDecrypt {
		log15.Warn("cannot run userCredentialMigrator.Down when decryption isn't allowed")
		return nil
	}

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}

	f := func() error {
		credentials, _, err := tx.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope: database.UserCredentialScope{
				Domain: database.UserCredentialDomainBatches,
			},
			LimitOffset: &database.LimitOffset{
				Limit: userCredentialMigrationCountPerRun,
			},
			ForUpdate:     true,
			OnlyEncrypted: true,
		})
		if err != nil {
			return errors.Wrap(err, "listing user credentials")
		}
		for _, cred := range credentials {
			a, err := cred.Authenticator(ctx)
			if err != nil {
				return errors.Wrapf(err, "retrieving authenticator for ID %d", cred.ID)
			}

			raw, err := database.EncryptAuthenticator(ctx, nil, a)
			if err != nil {
				return errors.Wrapf(err, "marshalling authenticator without an encrypter")
			}

			cred.EncryptedCredential = raw
			cred.EncryptionKeyID = database.UserCredentialUnmigratedEncryptionKeyID
			if err := tx.UserCredentials().Update(ctx, cred); err != nil {
				return errors.Wrapf(err, "upserting user credential %d", cred.ID)
			}
		}

		return nil
	}
	return tx.Done(f())
}
