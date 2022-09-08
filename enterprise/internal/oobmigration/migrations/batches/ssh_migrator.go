package batches

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type SSHMigrator struct {
	logger    log.Logger
	store     *basestore.Store
	key       encryption.Key
	batchSize int
}

var _ oobmigration.Migrator = &SSHMigrator{}

func NewSSHMigratorWithDB(store *basestore.Store, key encryption.Key, batchSize int) *SSHMigrator {
	return &SSHMigrator{
		logger:    log.Scoped("SSHMigrator", ""),
		store:     store,
		key:       key,
		batchSize: batchSize,
	}
}

func (m *SSHMigrator) ID() int                 { return 2 }
func (m *SSHMigrator) Interval() time.Duration { return time.Second * 5 }

// Progress returns the percentage (ranged [0, 1]) of external services without a marker
// indicating that this migration has been applied to that row.
func (m *SSHMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(sshMigratorProgressQuery, "batches", "batches")))
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

type jsonSSHMigratorAuth struct {
	Username string `json:"Username,omitempty"`
	Password string `json:"Password,omitempty"`
	Token    string `json:"Token,omitempty"`
}

type jsonSSHMigratorSSHFragment struct {
	PrivateKey string `json:"PrivateKey"`
	PublicKey  string `json:"PublicKey"`
	Passphrase string `json:"Passphrase"`
}

// Up generates a keypair for authenticators missing SSH credentials.
func (m *SSHMigrator) Up(ctx context.Context) (err error) {
	return m.run(ctx, false, func(credential string) (string, bool, error) {
		var envelope struct {
			Type    string          `json:"Type"`
			Payload json.RawMessage `json:"Auth"`
		}
		if err := json.Unmarshal([]byte(credential), &envelope); err != nil {
			return "", false, err
		}
		if envelope.Type != "BasicAuth" && envelope.Type != "OAuthBearerToken" {
			// Not a key type that supports SSH additions, leave credentials as-is
			return "", false, nil
		}

		auth := jsonSSHMigratorAuth{}
		if err := json.Unmarshal(envelope.Payload, &auth); err != nil {
			return "", false, err
		}

		keypair, err := encryption.GenerateRSAKey()
		if err != nil {
			return "", false, err
		}

		encoded, err := json.Marshal(struct {
			Type string
			Auth any
		}{
			Type: envelope.Type + "WithSSH",
			Auth: struct {
				jsonSSHMigratorAuth
				jsonSSHMigratorSSHFragment
			}{
				auth,
				jsonSSHMigratorSSHFragment{
					PrivateKey: keypair.PrivateKey,
					PublicKey:  keypair.PublicKey,
					Passphrase: keypair.Passphrase,
				},
			},
		})
		if err != nil {
			return "", false, err
		}

		return string(encoded), true, nil
	})
}

// Down converts all credentials with an SSH key back to a historically supported version.
func (m *SSHMigrator) Down(ctx context.Context) (err error) {
	return m.run(ctx, true, func(credential string) (string, bool, error) {
		var envelope struct {
			Type    string          `json:"Type"`
			Payload json.RawMessage `json:"Auth"`
		}
		if err := json.Unmarshal([]byte(credential), &envelope); err != nil {
			return "", false, err
		}
		if envelope.Type != "BasicAuthWithSSH" && envelope.Type != "OAuthBearerTokenWithSSH" {
			// Not a key type that that has SSH additions (nothing to remove)
			return "", false, nil
		}

		auth := jsonSSHMigratorAuth{}
		if err := json.Unmarshal(envelope.Payload, &auth); err != nil {
			return "", false, err
		}

		encoded, err := json.Marshal(struct {
			Type string
			Auth jsonSSHMigratorAuth
		}{
			Type: strings.TrimSuffix(envelope.Type, "WithSSH"),
			Auth: auth,
		})
		if err != nil {
			return "", false, err
		}

		return string(encoded), true, nil
	})
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
		rows, err := tx.Query(ctx, sqlf.Sprintf(sshMigratorSelectQuery, "batches", sshMigrationsApplied, m.batchSize))
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
			if err := tx.Exec(ctx, sqlf.Sprintf(sshMigratorUpdateFlagonlyQuery, !sshMigrationsApplied, credential.ID)); err != nil {
				return err
			}

			continue
		}

		secret, keyID, err := encryption.MaybeEncrypt(ctx, m.key, newCred)
		if err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(sshMigratorUpdateQuery, !sshMigrationsApplied, secret, keyID, credential.ID)); err != nil {
			return err
		}
	}

	return nil
}

const sshMigratorSelectQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:run
SELECT
	id,
	credential,
	encryption_key_id
FROM user_credentials
WHERE
	domain = %s AND
	ssh_migration_applied = %s
ORDER BY ID
LIMIT %s
FOR UPDATE
`

const sshMigratorUpdateQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:run
UPDATE user_credentials
SET
	updated_at = NOW(),
	ssh_migration_applied = %s,
	credential = %s,
	encryption_key_id = %s
WHERE id = %s
`

const sshMigratorUpdateFlagonlyQuery = `
-- source: enterprise/internal/oobmigration/migrations/batches/ssh_migrator.go:run
UPDATE user_credentials
SET ssh_migration_applied = %s
WHERE id = %s
`
