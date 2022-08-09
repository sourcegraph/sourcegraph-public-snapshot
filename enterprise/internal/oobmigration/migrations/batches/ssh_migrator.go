package batches

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SSHMigrator struct {
	logger    log.Logger
	store     *store.Store
	BatchSize int
	key       encryption.Key
}

var _ oobmigration.Migrator = &SSHMigrator{}

func NewSSHMigratorWithDB(store *store.Store /* db database.DB */) *SSHMigrator {
	return &SSHMigrator{
		logger:    log.Scoped("SSHMigrator", ""),
		store:     store, /* basestore.NewWithHandle(db.Handle()) */
		BatchSize: 5,
		key:       keyring.Default().BatchChangesCredentialKey, // TODO
	}
}

// Progress returns the percentage (ranged [0, 1]) of external services without a marker
// indicating that this migration has been applied to that row.
func (m *SSHMigrator) Progress(ctx context.Context) (float64, error) {
	domain := database.UserCredentialDomainBatches
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(sshMigratorProgressQuery, domain, domain)))
	return progress, err
}

const sshMigratorProgressQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:Progress
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST((c2.count - c1.count) AS float) / CAST(c2.count AS float)
	END
FROM
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s AND NOT ssh_migration_applied) c1,
	(SELECT COUNT(*) as count FROM user_credentials WHERE domain = %s) c2
`

// Up generates a keypair for authenticators missing SSH credentials.
func (m *SSHMigrator) Up(ctx context.Context) (err error) {
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
			Limit: m.BatchSize,
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

		var newCred auth.Authenticator
		switch a := a.(type) {
		case *auth.OAuthBearerToken:
			cred := &auth.OAuthBearerTokenWithSSH{OAuthBearerToken: *a}
			keypair, err := encryption.GenerateRSAKey()
			if err != nil {
				return err
			}
			cred.PrivateKey = keypair.PrivateKey
			cred.PublicKey = keypair.PublicKey
			cred.Passphrase = keypair.Passphrase
			newCred = cred

		case *auth.BasicAuth:
			cred := &auth.BasicAuthWithSSH{BasicAuth: *a}
			keypair, err := encryption.GenerateRSAKey()
			if err != nil {
				return err
			}
			cred.PrivateKey = keypair.PrivateKey
			cred.PublicKey = keypair.PublicKey
			cred.Passphrase = keypair.Passphrase
			newCred = cred
		}
		if newCred != nil {
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

// Down converts all credentials with an SSH key back to a historically supported version.
func (m *SSHMigrator) Down(ctx context.Context) (err error) {
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
			Limit: m.BatchSize,
		},
		ForUpdate:           true,
		SSHMigrationApplied: &t,
	})
	for _, cred := range credentials {
		a, err := cred.Authenticator(ctx)
		if err != nil {
			return err
		}

		var newCred auth.Authenticator
		switch a := a.(type) {
		case *auth.OAuthBearerTokenWithSSH:
			newCred = &a.OAuthBearerToken
		case *auth.BasicAuthWithSSH:
			newCred = &a.BasicAuth
		}
		if newCred != nil {
			secret, id, err := database.EncryptAuthenticator(ctx, m.key, newCred)
			if err != nil {
				return errors.Wrap(err, "encrypting authenticator")
			}

			cred.EncryptedCredential = secret
			cred.EncryptionKeyID = id
		}

		cred.SSHMigrationApplied = false
		if err := tx.UserCredentials().Update(ctx, cred); err != nil {
			return err
		}
	}

	return nil
}
