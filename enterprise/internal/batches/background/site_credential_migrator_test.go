package background

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestSiteCredentialMigrator(t *testing.T) {
	ctx := context.Background()
	db := dbtesting.GetDB(t)

	cstore := store.New(db, et.TestKey{})

	migrator := &siteCredentialMigrator{cstore}
	a := &auth.BasicAuth{Username: "foo", Password: "bar"}

	t.Run("no user credentials", func(t *testing.T) {
		assertProgress(t, ctx, 1.0, migrator)
	})

	// Now we'll set up enough users to validate that it takes multiple Up
	// invocations.
	for i := 0; i < siteCredentialMigrationCountPerRun*2; i++ {
		cred := &btypes.SiteCredential{
			ExternalServiceType: extsvc.TypeGitLab,
			ExternalServiceID:   fmt.Sprintf("https://%d.gitlab.com/", i),
		}
		if err := cstore.CreateSiteCredential(ctx, cred, a); err != nil {
			t.Fatal(err)
		}

		// Override the saved credential to only include the unencrypted
		// authenticator.
		cred.SetRawCredential(a, nil)
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
