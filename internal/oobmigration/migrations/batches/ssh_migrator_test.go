package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
)

func TestSSHMigrator(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := basestore.NewWithHandle(db.Handle())
	key := et.TestKey{}

	userID, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`
		INSERT INTO users (username, display_name, created_at)
		VALUES (%s, %s, NOW())
		RETURNING id
	`,
		"testuser-0",
		"testuser",
	)))
	if err != nil {
		t.Fatal(err)
	}

	encryption.MockGenerateRSAKey = func() (key *encryption.RSAKey, err error) {
		return &encryption.RSAKey{
			PrivateKey: "private",
			Passphrase: "pass",
			PublicKey:  "public",
		}, nil
	}
	t.Cleanup(func() {
		encryption.MockGenerateRSAKey = nil
	})

	migrator := NewSSHMigratorWithDB(store, key, 5)
	progress, err := migrator.Progress(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	credential, keyID, err := encryption.MaybeEncrypt(ctx, key, `{"type": "OAuthBearerToken", "auth": {"token": "test"}}`)
	if err != nil {
		t.Fatal(err)
	}

	credentialID, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`
		INSERT INTO user_credentials (domain, user_id, external_service_type, external_service_id, credential, encryption_key_id, ssh_migration_applied)
		VALUES (%s, %s, %s, %s, %s, %s, false)
		RETURNING id
	`,
		"batches",
		userID,
		"GITHUB",
		"https://github.com/",
		credential,
		keyID,
	)))
	if err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := progress, 0.0; have != want {
		t.Fatalf("got invalid progress with one unmigrated entry, want=%f have=%f", want, have)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	{
		migratedCredential, ok, err := scanFirstCredential(store.Query(ctx, sqlf.Sprintf(`
			SELECT id, domain, user_id, external_service_type, external_service_id, ssh_migration_applied, credential, encryption_key_id
			FROM user_credentials WHERE id = %s
		`,
			credentialID,
		)))
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("no credential")
		}

		if have, want := migratedCredential.domain, "batches"; have != want {
			t.Fatalf("invalid Domain after migration, want=%q have=%q", want, have)
		}
		if have, want := migratedCredential.userID, int32(userID); have != want {
			t.Fatalf("invalid UserID after migration, want=%d have=%d", want, have)
		}
		if have, want := migratedCredential.externalServiceType, "GITHUB"; have != want {
			t.Fatalf("invalid ExternalServiceType after migration, want=%q have=%q", want, have)
		}
		if have, want := migratedCredential.externalServiceID, "https://github.com/"; have != want {
			t.Fatalf("invalid ExternalServiceID after migration, want=%q have=%q", want, have)
		}
		if !migratedCredential.sshMigrationApplied {
			t.Fatalf("invalid migration flag: have=%v want=%v", migratedCredential.sshMigrationApplied, true)
		}

		decrypted, err := encryption.MaybeDecrypt(ctx, key, migratedCredential.encryptedConfig, migratedCredential.keyID)
		if err != nil {
			t.Fatal(err)
		}
		var credential struct {
			Type string
			Auth struct {
				Token      string
				PrivateKey string
				PublicKey  string
				Passphrase string
			}
		}
		if err := json.Unmarshal([]byte(decrypted), &credential); err != nil {
			t.Fatal(err)
		}
		if credential.Type != "OAuthBearerTokenWithSSH" {
			t.Fatalf("invalid type of migrated credential: %s", credential.Type)
		}
		if have, want := credential.Auth.Token, "test"; have != want {
			t.Fatalf("invalid token stored in migrated credential, want=%q have=%q", want, have)
		}
		if credential.Auth.Passphrase == "" || credential.Auth.PrivateKey == "" || credential.Auth.PublicKey == "" {
			t.Fatal("ssh keypair is not complete")
		}
	}

	if err := migrator.Down(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := progress, 0.0; have != want {
		t.Fatalf("got invalid progress after down migration, want=%f have=%f", want, have)
	}

	{
		migratedCredential, ok, err := scanFirstCredential(store.Query(ctx, sqlf.Sprintf(`
			SELECT id, domain, user_id, external_service_type, external_service_id, ssh_migration_applied, credential, encryption_key_id
			FROM user_credentials WHERE id = %s
		`,
			credentialID,
		)))
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("no credential")
		}

		if have, want := migratedCredential.domain, "batches"; have != want {
			t.Fatalf("invalid Domain after down migration, want=%q have=%q", want, have)
		}
		if have, want := migratedCredential.userID, int32(userID); have != want {
			t.Fatalf("invalid UserID after down migration, want=%d have=%d", want, have)
		}
		if have, want := migratedCredential.externalServiceType, "GITHUB"; have != want {
			t.Fatalf("invalid ExternalServiceType after down migration, want=%q have=%q", want, have)
		}
		if have, want := migratedCredential.externalServiceID, "https://github.com/"; have != want {
			t.Fatalf("invalid ExternalServiceID after down migration, want=%q have=%q", want, have)
		}
		if migratedCredential.sshMigrationApplied {
			t.Fatalf("invalid migration flag: have=%v want=%v", migratedCredential.sshMigrationApplied, false)
		}

		decrypted, err := encryption.MaybeDecrypt(ctx, key, migratedCredential.encryptedConfig, migratedCredential.keyID)
		if err != nil {
			t.Fatal(err)
		}
		var credential struct {
			Type string
			Auth struct {
				Token string
			}
		}
		if err := json.Unmarshal([]byte(decrypted), &credential); err != nil {
			t.Fatal(err)
		}
		if credential.Type != "OAuthBearerToken" {
			t.Fatalf("invalid type of migrated credential: %s", credential.Type)
		}
		if have, want := credential.Auth.Token, "test"; have != want {
			t.Fatalf("invalid token stored in migrated credential, want=%q have=%q", want, have)
		}
	}
}

type userCredential struct {
	id                  int64
	domain              string
	userID              int32
	externalServiceType string
	externalServiceID   string
	sshMigrationApplied bool
	encryptedConfig     string
	keyID               string
}

var scanFirstCredential = basestore.NewFirstScanner(func(s dbutil.Scanner) (uc userCredential, _ error) {
	err := s.Scan(&uc.id, &uc.domain, &uc.userID, &uc.externalServiceType, &uc.externalServiceID, &uc.sshMigrationApplied, &uc.encryptedConfig, &uc.keyID)
	return uc, err
})
