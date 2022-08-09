package batches

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SSHMigrator struct {
	logger    log.Logger
	store     *basestore.Store
	BatchSize int
	key       encryption.Key
}

var _ oobmigration.Migrator = &SSHMigrator{}

func NewSSHMigratorWithDB(db database.DB, key encryption.Key) *SSHMigrator {
	return &SSHMigrator{
		logger:    log.Scoped("SSHMigrator", ""),
		store:     basestore.NewWithHandle(db.Handle()),
		BatchSize: 5,
		key:       key,
	}
}

func (m *SSHMigrator) ID() int {
	return 2
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
	transformer := func(a auth.Authenticator) (auth.Authenticator, bool, error) {
		keypair, err := encryption.GenerateRSAKey()
		if err != nil {
			return nil, false, err
		}

		switch a := a.(type) {
		case *auth.OAuthBearerToken:
			return &auth.OAuthBearerTokenWithSSH{
				OAuthBearerToken: *a,
				PrivateKey:       keypair.PrivateKey,
				PublicKey:        keypair.PublicKey,
				Passphrase:       keypair.Passphrase,
			}, true, nil

		case *auth.BasicAuth:
			return &auth.BasicAuthWithSSH{
				BasicAuth:  *a,
				PrivateKey: keypair.PrivateKey,
				PublicKey:  keypair.PublicKey,
				Passphrase: keypair.Passphrase,
			}, true, nil

		default:
			return nil, false, nil
		}
	}

	return m.run(ctx, false, transformer)
}

// Down converts all credentials with an SSH key back to a historically supported version.
func (m *SSHMigrator) Down(ctx context.Context) (err error) {
	transformer := func(a auth.Authenticator) (auth.Authenticator, bool, error) {
		switch a := a.(type) {
		case *auth.OAuthBearerTokenWithSSH:
			return &a.OAuthBearerToken, true, nil
		case *auth.BasicAuthWithSSH:
			return &a.BasicAuth, true, nil
		default:
			return nil, false, err
		}
	}

	return m.run(ctx, true, transformer)
}

func (m *SSHMigrator) run(ctx context.Context, sshMigrationsApplied bool, f func(auth.Authenticator) (auth.Authenticator, bool, error)) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	type credential struct {
		ID         int
		Credential string
	}
	credentials, err := func() (credentials []credential, err error) {
		rows, err := tx.Query(ctx, sqlf.Sprintf(
			sshMigratorSelectQuery,
			database.UserCredentialDomainBatches,
			sshMigrationsApplied,
			m.BatchSize,
		))
		if err != nil {
			return nil, err
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			var id int
			var rawCredential, keyID string
			if err := rows.Scan(&id, &rawCredential, &keyID); err != nil {
				return nil, err
			}
			if keyID != "" {
				decrypted, err := encryption.MaybeDecrypt(ctx, m.key, rawCredential, keyID)
				if err != nil {
					return nil, err
				}

				rawCredential = decrypted
			}

			credentials = append(credentials, credential{ID: id, Credential: rawCredential})
		}

		return credentials, nil
	}()
	if err != nil {
		return err
	}

	for _, credential := range credentials {
		secret, id, ok, err := m.transform(ctx, credential.Credential, f)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(sshMigratorUpdateQuery, secret, id, !sshMigrationsApplied, credential.ID)); err != nil {
			return err
		}
	}

	return nil
}

const sshMigratorSelectQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:run
SELECT id, credential, encryption_key_id FROM user_credentials WHERE domain = %s AND ssh_migration_applied = %s ORDER BY ID LIMIT %s FOR UPDATE
`

const sshMigratorUpdateQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:run
UPDATE user_credentials
SET
	credential = %s,
	encryption_key_id = %s,
	updated_at = NOW(),
	ssh_migration_applied = %s
WHERE id = %s
`

func (m *SSHMigrator) transform(ctx context.Context, credential string, f func(auth.Authenticator) (auth.Authenticator, bool, error)) ([]byte, string, bool, error) {
	a, err := database.UnmarshalAuthenticator(credential)
	if err != nil {
		return nil, "", false, errors.Wrap(err, "unmarshalling authenticator")
	}

	newCred, ok, err := f(a)
	if err != nil {
		return nil, "", false, err
	}
	if !ok {
		return nil, "", false, nil
	}

	secret, id, err := database.EncryptAuthenticator(ctx, m.key, newCred)
	if err != nil {
		return nil, "", false, errors.Wrap(err, "encrypting authenticator")
	}

	return secret, id, true, nil
}
