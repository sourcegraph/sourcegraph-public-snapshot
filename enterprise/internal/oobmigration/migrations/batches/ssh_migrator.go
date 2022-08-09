package batches

import (
	"context"
	"encoding/json"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
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
	transformer := func(credential string) (string, bool, error) {
		var partial struct {
			Type database.AuthenticatorType
			Auth json.RawMessage
		}
		if err := json.Unmarshal([]byte(credential), &partial); err != nil {
			return "", false, err
		}
		var a auth.Authenticator
		switch partial.Type {
		case database.AuthenticatorTypeOAuthClient:
			a = &auth.OAuthClient{}
		case database.AuthenticatorTypeBasicAuth:
			a = &auth.BasicAuth{}
		case database.AuthenticatorTypeBasicAuthWithSSH:
			a = &auth.BasicAuthWithSSH{}
		case database.AuthenticatorTypeOAuthBearerToken:
			a = &auth.OAuthBearerToken{}
		case database.AuthenticatorTypeOAuthBearerTokenWithSSH:
			a = &auth.OAuthBearerTokenWithSSH{}
		case database.AuthenticatorTypeBitbucketServerSudoableOAuthClient:
			a = &bitbucketserver.SudoableOAuthClient{}
		case database.AuthenticatorTypeGitLabSudoableToken:
			a = &gitlab.SudoableToken{}
		default:
			return "", false, errors.Errorf("unknown credential type: %s", partial.Type)
		}
		if err := json.Unmarshal(partial.Auth, &a); err != nil {
			return "", false, err
		}

		keypair, err := encryption.GenerateRSAKey()
		if err != nil {
			return "", false, err
		}

		switch a := a.(type) {
		case *auth.OAuthBearerToken:
			rawx, err := json.Marshal(struct {
				Type string
				Auth auth.Authenticator
			}{
				Type: "OAuthBearerTokenWithSSH",
				Auth: &auth.OAuthBearerTokenWithSSH{
					OAuthBearerToken: *a,
					PrivateKey:       keypair.PrivateKey,
					PublicKey:        keypair.PublicKey,
					Passphrase:       keypair.Passphrase,
				},
			})
			if err != nil {
				return "", false, err
			}

			return string(rawx), true, nil

		case *auth.BasicAuth:
			rawx, err := json.Marshal(struct {
				Type string
				Auth auth.Authenticator
			}{
				Type: "BasicAuthWithSSH",
				Auth: &auth.BasicAuthWithSSH{
					BasicAuth:  *a,
					PrivateKey: keypair.PrivateKey,
					PublicKey:  keypair.PublicKey,
					Passphrase: keypair.Passphrase,
				},
			})
			if err != nil {
				return "", false, err
			}

			return string(rawx), true, nil

		default:
			return "", false, nil
		}
	}

	return m.run(ctx, false, transformer)
}

// Down converts all credentials with an SSH key back to a historically supported version.
func (m *SSHMigrator) Down(ctx context.Context) (err error) {
	transformer := func(credential string) (string, bool, error) {
		var partial struct {
			Type database.AuthenticatorType
			Auth json.RawMessage
		}
		if err := json.Unmarshal([]byte(credential), &partial); err != nil {
			return "", false, err
		}
		var a auth.Authenticator
		switch partial.Type {
		case database.AuthenticatorTypeOAuthClient:
			a = &auth.OAuthClient{}
		case database.AuthenticatorTypeBasicAuth:
			a = &auth.BasicAuth{}
		case database.AuthenticatorTypeBasicAuthWithSSH:
			a = &auth.BasicAuthWithSSH{}
		case database.AuthenticatorTypeOAuthBearerToken:
			a = &auth.OAuthBearerToken{}
		case database.AuthenticatorTypeOAuthBearerTokenWithSSH:
			a = &auth.OAuthBearerTokenWithSSH{}
		case database.AuthenticatorTypeBitbucketServerSudoableOAuthClient:
			a = &bitbucketserver.SudoableOAuthClient{}
		case database.AuthenticatorTypeGitLabSudoableToken:
			a = &gitlab.SudoableToken{}
		default:
			return "", false, errors.Errorf("unknown credential type: %s", partial.Type)
		}
		if err := json.Unmarshal(partial.Auth, &a); err != nil {
			return "", false, err
		}

		switch a := a.(type) {
		case *auth.OAuthBearerTokenWithSSH:
			rawx, err := json.Marshal(struct {
				Type string
				Auth any
			}{
				Type: "OAuthBearerToken",
				Auth: a.OAuthBearerToken,
			})
			if err != nil {
				return "", false, err
			}

			return string(rawx), true, nil

		case *auth.BasicAuthWithSSH:
			rawx, err := json.Marshal(struct {
				Type string
				Auth any
			}{
				Type: "BasicAuth",
				Auth: a.BasicAuth,
			})
			if err != nil {
				return "", false, err
			}

			return string(rawx), true, nil

		default:
			return "", false, err
		}
	}

	return m.run(ctx, true, transformer)
}

func (m *SSHMigrator) run(ctx context.Context, sshMigrationsApplied bool, f func(string) (string, bool, error)) (err error) {
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
		rows, err := tx.Query(ctx, sqlf.Sprintf(sshMigratorSelectQuery, database.UserCredentialDomainBatches, sshMigrationsApplied, m.BatchSize))
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
			rawCredential, err = encryption.MaybeDecrypt(ctx, m.key, rawCredential, keyID)
			if err != nil {
				return nil, err
			}

			credentials = append(credentials, credential{ID: id, Credential: rawCredential})
		}

		return credentials, nil
	}()
	if err != nil {
		return err
	}

	for _, credential := range credentials {
		newCred, ok, err := f(credential.Credential)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		secret, keyID, err := encryption.MaybeEncrypt(ctx, m.key, newCred)
		if err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(sshMigratorUpdateQuery, secret, keyID, !sshMigrationsApplied, credential.ID)); err != nil {
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
