package background

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestUserCredentialMigrator(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t, "")

	cstore := store.New(db, et.TestKey{})

	migrator := &userCredentialMigrator{cstore, true}
	a := &auth.BasicAuth{Username: "foo", Password: "bar"}

	t.Run("no user credentials", func(t *testing.T) {
		assertProgress(t, ctx, 1.0, migrator)
	})

	// Now we'll set up enough users to validate that it takes multiple Up
	// invocations.
	for i := 0; i < userCredentialMigrationCountPerRun; i++ {
		user := ct.CreateTestUser(t, db, false)
		createUnencryptedUserCredential(t, ctx, cstore, database.UserCredentialScope{
			Domain:              database.UserCredentialDomainBatches,
			UserID:              user.ID,
			ExternalServiceType: extsvc.TypeGitLab,
			ExternalServiceID:   "https://gitlab.com/",
		}, a)
	}
	for i := 0; i < userCredentialMigrationCountPerRun; i++ {
		user := ct.CreateTestUser(t, db, false)
		createPreviouslyEncryptedUserCredential(t, ctx, cstore, database.UserCredentialScope{
			Domain:              database.UserCredentialDomainBatches,
			UserID:              user.ID,
			ExternalServiceType: extsvc.TypeGitLab,
			ExternalServiceID:   "https://gitlab.com/",
		}, a)
	}

	t.Run("completely unmigrated", func(t *testing.T) {
		assertProgress(t, ctx, 0.0, migrator)
	})

	t.Run("first migrate up", func(t *testing.T) {
		if err := migrator.Up(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assertProgress(t, ctx, 0.5, migrator)
	})

	t.Run("second migrate up", func(t *testing.T) {
		if err := migrator.Up(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assertProgress(t, ctx, 1.0, migrator)
	})

	t.Run("check credentials", func(t *testing.T) {
		credentials, _, err := cstore.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope: database.UserCredentialScope{Domain: database.UserCredentialDomainBatches},
		})
		if err != nil {
			t.Fatal(err)
		}

		for _, cred := range credentials {
			have, err := cred.Authenticator(ctx)
			if err != nil {
				t.Logf("cred: %+v", cred)
				t.Errorf("cannot get authenticator: %v", err)
			}

			if diff := cmp.Diff(have, a); diff != "" {
				t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
			}
		}

		// Finally, let's ensure there's nothing left to be migrated.
		if creds, _, err := cstore.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope:             database.UserCredentialScope{Domain: database.UserCredentialDomainBatches},
			RequiresMigration: true,
		}); err != nil {
			t.Fatal(err)
		} else if len(creds) > 0 {
			t.Errorf("unexpected unmigrated user credentials: %d", len(creds))
		}
	})

	t.Run("migrate down without allowing", func(t *testing.T) {
		migrator.allowDecrypt = false
		t.Cleanup(func() { migrator.allowDecrypt = true })

		if err := migrator.Down(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Nothing should have changed.
		assertProgress(t, ctx, 1.0, migrator)
	})

	t.Run("first migrate down", func(t *testing.T) {
		if err := migrator.Down(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assertProgress(t, ctx, 0.5, migrator)
	})

	t.Run("second migrate down", func(t *testing.T) {
		if err := migrator.Down(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assertProgress(t, ctx, 0.0, migrator)
	})

	t.Run("check credentials", func(t *testing.T) {
		credentials, _, err := cstore.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope: database.UserCredentialScope{Domain: database.UserCredentialDomainBatches},
		})
		if err != nil {
			t.Fatal(err)
		}

		for _, cred := range credentials {
			have, err := cred.Authenticator(ctx)
			if err != nil {
				t.Logf("cred: %+v", cred)
				t.Errorf("cannot get authenticator: %v", err)
			}

			if diff := cmp.Diff(have, a); diff != "" {
				t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
			}
		}

		// Finally, let's ensure there's nothing left to be migrated.
		if creds, _, err := cstore.UserCredentials().List(ctx, database.UserCredentialsListOpts{
			Scope:             database.UserCredentialScope{Domain: database.UserCredentialDomainBatches},
			RequiresMigration: true,
		}); err != nil {
			t.Fatal(err)
		} else if want := siteCredentialMigrationCountPerRun * 2; len(creds) != want {
			t.Errorf("unexpected number of unencrypted credentials: have=%d want=%d", len(creds), want)
		}
	})
}

func assertProgress(t *testing.T, ctx context.Context, want float64, migrator interface {
	Progress(context.Context) (float64, error)
}) {
	t.Helper()

	if have, err := migrator.Progress(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if have != want {
		t.Errorf("unexpected progress: have=%f want=%f", have, want)
	}
}

func createUnencryptedUserCredential(
	t *testing.T,
	ctx context.Context,
	store *store.Store,
	scope database.UserCredentialScope,
	a auth.Authenticator,
) *database.UserCredential {
	cred, err := store.UserCredentials().Create(ctx, scope, a)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(struct {
		Type database.AuthenticatorType
		Auth auth.Authenticator
	}{
		Type: database.AuthenticatorTypeBasicAuth,
		Auth: a,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Exec(
		ctx,
		sqlf.Sprintf(
			"UPDATE user_credentials SET credential = %s, encryption_key_id = %s WHERE id = %s",
			raw,
			database.UserCredentialUnmigratedEncryptionKeyID,
			cred.ID,
		),
	); err != nil {
		t.Fatal(err)
	}

	return cred
}

func createPreviouslyEncryptedUserCredential(
	t *testing.T,
	ctx context.Context,
	store *store.Store,
	scope database.UserCredentialScope,
	a auth.Authenticator,
) *database.UserCredential {
	cred, err := store.UserCredentials().Create(ctx, scope, a)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Exec(
		ctx,
		sqlf.Sprintf(
			"UPDATE user_credentials SET encryption_key_id = %s WHERE id = %s",
			database.UserCredentialPlaceholderEncryptionKeyID,
			cred.ID,
		),
	); err != nil {
		t.Fatal(err)
	}

	return cred
}
