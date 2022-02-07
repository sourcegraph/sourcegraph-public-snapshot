package migrations

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const siteCredentialMigrationCountPerRun = 5

type siteCredentialMigrator struct {
	store        *store.Store
	allowDecrypt bool
}

var _ oobmigration.Migrator = &siteCredentialMigrator{}

func (m *siteCredentialMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(
		m.store.Query(ctx, sqlf.Sprintf(
			siteCredentialMigratorProgressQuery,
			btypes.SiteCredentialPlaceholderEncryptionKeyID,
			btypes.SiteCredentialUnmigratedEncryptionKeyID,
		)),
	)
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const siteCredentialMigratorProgressQuery = `
-- source: enterprise/internal/batches/site_credential_migrator.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE CAST((c2.count - c1.count) AS float) / CAST(c2.count AS float) END FROM
	(SELECT COUNT(*) as count FROM batch_changes_site_credentials WHERE encryption_key_id IN (%s, %s)) c1,
	(SELECT COUNT(*) as count FROM batch_changes_site_credentials) c2
`

func (m *siteCredentialMigrator) Up(ctx context.Context) error {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}

	f := func() error {
		credentials, _, err := tx.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{
			LimitOpts:         store.LimitOpts{Limit: siteCredentialMigrationCountPerRun},
			RequiresMigration: true,
			ForUpdate:         true,
		})
		if err != nil {
			return errors.Wrap(err, "listing site credentials")
		}
		for _, cred := range credentials {
			a, err := cred.Authenticator(ctx)
			if err != nil {
				return errors.Wrapf(err, "retrieving authenticator for ID %d", cred.ID)
			}

			if err := cred.SetAuthenticator(ctx, a); err != nil {
				return errors.Wrapf(err, "setting authenticator for ID %d", cred.ID)
			}

			if err := tx.UpdateSiteCredential(ctx, cred); err != nil {
				return errors.Wrapf(err, "updating site credential %d", cred.ID)
			}
		}

		return nil
	}
	return tx.Done(f())
}

func (m *siteCredentialMigrator) Down(ctx context.Context) error {
	if !m.allowDecrypt {
		log15.Warn("cannot run siteCredentialMigrator.Down when decryption isn't allowed")
		return nil
	}

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}

	f := func() error {
		credentials, _, err := tx.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{
			LimitOpts:     store.LimitOpts{Limit: siteCredentialMigrationCountPerRun},
			OnlyEncrypted: true,
			ForUpdate:     true,
		})
		if err != nil {
			return errors.Wrap(err, "listing site credentials")
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
			cred.EncryptionKeyID = btypes.SiteCredentialUnmigratedEncryptionKeyID
			if err := tx.UpdateSiteCredential(ctx, cred); err != nil {
				return errors.Wrapf(err, "upserting site credential %d", cred.ID)
			}
		}

		return nil
	}
	return tx.Done(f())
}
