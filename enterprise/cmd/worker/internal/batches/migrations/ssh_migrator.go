package migrations

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

const sshMigrationCountPerRun = 5

// sshMigrator migrates existing batch changes credentials that have no SSH key stored
// to a variant that includes it.
type sshMigrator struct {
	store *store.Store
}

var _ oobmigration.Migrator = &sshMigrator{}

// Progress returns the ratio of migrated records to total records. Any record with a
// credential type that ends on WithSSH is considered migrated.
func (m *sshMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(
		m.store.Query(ctx, sqlf.Sprintf(
			sshMigratorProgressQuery,
			database.UserCredentialDomainBatches,
			database.UserCredentialDomainBatches,
		)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const sshMigratorProgressQuery = `
-- source: enterprise/internal/batches/ssh_migrator.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE CAST((c2.count - c1.count) AS float) / CAST(c2.count AS float) END FROM
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s AND ssh_migration_applied = FALSE) c1,
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s) c2
`

// Up loops over all credentials and finds authenticators that are missing
// SSH credentials, generates a keypair for them and upgrades them.
func (m *sshMigrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	f := false
	credentials, _, err := tx.UserCredentials().List(ctx, database.UserCredentialsListOpts{
		Scope: database.UserCredentialScope{
			Domain: database.UserCredentialDomainBatches,
		},
		LimitOffset: &database.LimitOffset{
			Limit: sshMigrationCountPerRun,
		},
		ForUpdate:           true,
		SSHMigrationApplied: &f,
	})
	if err != nil {
		return err
	}
	for _, cred := range credentials {
		a, err := cred.Authenticator(ctx)
		if err != nil {
			return err
		}

		switch a := a.(type) {
		case *auth.OAuthBearerToken:
			newCred := &auth.OAuthBearerTokenWithSSH{OAuthBearerToken: *a}
			keypair, err := encryption.GenerateRSAKey()
			if err != nil {
				return err
			}
			newCred.PrivateKey = keypair.PrivateKey
			newCred.PublicKey = keypair.PublicKey
			newCred.Passphrase = keypair.Passphrase
			if err := cred.SetAuthenticator(ctx, newCred); err != nil {
				return err
			}
		case *auth.BasicAuth:
			newCred := &auth.BasicAuthWithSSH{BasicAuth: *a}
			keypair, err := encryption.GenerateRSAKey()
			if err != nil {
				return err
			}
			newCred.PrivateKey = keypair.PrivateKey
			newCred.PublicKey = keypair.PublicKey
			newCred.Passphrase = keypair.Passphrase
			if err := cred.SetAuthenticator(ctx, newCred); err != nil {
				return err
			}
		}

		cred.SSHMigrationApplied = true
		if err := tx.UserCredentials().Update(ctx, cred); err != nil {
			return err
		}
	}

	return nil
}

// Down converts all credentials that have an SSH key back to a version without, so
// an older version of Sourcegraph would be able to match those credentials again.
func (m *sshMigrator) Down(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	t := true
	credentials, _, err := tx.UserCredentials().List(ctx, database.UserCredentialsListOpts{
		Scope: database.UserCredentialScope{
			Domain: database.UserCredentialDomainBatches,
		},
		LimitOffset: &database.LimitOffset{
			Limit: sshMigrationCountPerRun,
		},
		ForUpdate:           true,
		SSHMigrationApplied: &t,
	})
	for _, cred := range credentials {
		a, err := cred.Authenticator(ctx)
		if err != nil {
			return err
		}

		switch a := a.(type) {
		case *auth.OAuthBearerTokenWithSSH:
			newCred := &a.OAuthBearerToken
			if err := cred.SetAuthenticator(ctx, newCred); err != nil {
				return err
			}
		case *auth.BasicAuthWithSSH:
			newCred := &a.BasicAuth
			if err := cred.SetAuthenticator(ctx, newCred); err != nil {
				return err
			}
		}

		cred.SSHMigrationApplied = false
		if err := tx.UserCredentials().Update(ctx, cred); err != nil {
			return err
		}
	}

	return nil
}
