package background

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestSiteCredentialMigrator(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t, "")

	cstore := store.New(db, et.TestKey{})

	migrator := &siteCredentialMigrator{cstore, true}
	a := &auth.BasicAuth{Username: "foo", Password: "bar"}

	t.Run("no user credentials", func(t *testing.T) {
		assertProgress(t, ctx, 1.0, migrator)
	})

	// Now we'll set up enough users to validate that it takes multiple Up
	// invocations.
	for i := 0; i < siteCredentialMigrationCountPerRun; i++ {
		cred := &btypes.SiteCredential{
			ExternalServiceType: extsvc.TypeGitLab,
			ExternalServiceID:   fmt.Sprintf("https://%d.gitlab.com/", i),
		}
		if err := cstore.CreateSiteCredential(ctx, cred, a); err != nil {
			t.Fatal(err)
		}

		// Override the saved credential to only include the unencrypted
		// authenticator.
		enc, err := database.EncryptAuthenticator(ctx, nil, a)
		if err != nil {
			t.Fatal(err)
		}

		cred.EncryptedCredential = enc
		cred.EncryptionKeyID = btypes.SiteCredentialUnmigratedEncryptionKeyID
		if err := cstore.UpdateSiteCredential(ctx, cred); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < siteCredentialMigrationCountPerRun; i++ {
		cred := &btypes.SiteCredential{
			ExternalServiceType: extsvc.TypeGitLab,
			ExternalServiceID:   fmt.Sprintf("https://%d.gitlab.com/", i+siteCredentialMigrationCountPerRun),
		}
		if err := cstore.CreateSiteCredential(ctx, cred, a); err != nil {
			t.Fatal(err)
		}

		// Override the saved credential to only include the placeholder
		// encryption key ID.
		cred.EncryptionKeyID = btypes.SiteCredentialPlaceholderEncryptionKeyID
		if err := cstore.UpdateSiteCredential(ctx, cred); err != nil {
			t.Fatal(err)
		}
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
		credentials, _, err := cstore.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{})
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
		if creds, _, err := cstore.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{
			RequiresMigration: true,
		}); err != nil {
			t.Fatal(err)
		} else if len(creds) > 0 {
			t.Errorf("unexpected number of unencrypted credentials: have=%d want=0", len(creds))
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
		credentials, _, err := cstore.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{})
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
		if creds, _, err := cstore.ListSiteCredentials(ctx, store.ListSiteCredentialsOpts{
			RequiresMigration: true,
		}); err != nil {
			t.Fatal(err)
		} else if want := siteCredentialMigrationCountPerRun * 2; len(creds) != want {
			t.Errorf("unexpected number of unencrypted credentials: have=%d want=%d", len(creds), want)
		}
	})
}
