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
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestUserCredentialMigrator(t *testing.T) {
	ctx := context.Background()
	db := dbtesting.GetDB(t)

	cstore := store.New(db)
	key := et.TestKey{}

	migrator := &userCredentialMigrator{cstore, key}
	a := &auth.BasicAuth{Username: "foo", Password: "bar"}

	t.Run("no user credentials", func(t *testing.T) {
		assertProgress(t, ctx, 1.0, migrator)
	})

	// Now we'll set up enough users to validate that it takes multiple Up
	// invocations.
	for i := 0; i < userCredentialMigrationCountPerRun*2; i++ {
		user := ct.CreateTestUser(t, db, false)
		createUnencryptedUserCredential(t, ctx, cstore, key, database.UserCredentialScope{
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
			have, err := cred.Authenticator(ctx, key)
			if err != nil {
				t.Errorf("cannot get authenticator: %v", err)
			}

			if diff := cmp.Diff(have, a); diff != "" {
				t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
			}
		}

		// Let's get down into the weeds and verify that there are no non-NULL
		// credential fields.
		if count, _, err := basestore.ScanFirstInt(cstore.Query(ctx, sqlf.Sprintf("SELECT COUNT(*) FROM user_credentials WHERE credential IS NOT NULL"))); err != nil {
			t.Errorf("cannot check unencrypted credentials: %v", err)
		} else if count != 0 {
			t.Errorf("unexpected number of unencrypted credentials: have=%d want=0", count)
		}
	})

	t.Run("down", func(t *testing.T) {
		if err := migrator.Down(ctx); err == nil {
			t.Error("unexpected nil error as down migrations are unsupported")
		}
	})
}

func assertProgress(t *testing.T, ctx context.Context, want float64, migrator *userCredentialMigrator) {
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
	key encryption.Key,
	scope database.UserCredentialScope,
	a auth.Authenticator) *database.UserCredential {
	cred, err := store.UserCredentials().Create(ctx, key, scope, a)
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
			"UPDATE user_credentials SET credential = %s, credential_enc = NULL WHERE id = %s",
			raw,
			cred.ID,
		),
	); err != nil {
		t.Fatal(err)
	}

	return cred
}
