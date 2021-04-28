package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func testStoreSiteCredentials(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	credentials := make([]*btypes.SiteCredential, 0, 3)
	// Make sure these are sorted alphabetically.
	externalServiceTypes := []string{
		extsvc.TypeBitbucketServer,
		extsvc.TypeGitHub,
		extsvc.TypeGitLab,
	}

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(credentials); i++ {
			cred := &btypes.SiteCredential{
				ExternalServiceType: externalServiceTypes[i],
				ExternalServiceID:   "https://someurl.test",
				Credential:          &auth.OAuthBearerToken{Token: "123"},
			}

			if err := s.CreateSiteCredential(ctx, cred); err != nil {
				t.Fatal(err)
			}
			if cred.ID == 0 {
				t.Fatal("id should not be zero")
			}
			if cred.CreatedAt.IsZero() {
				t.Fatal("CreatedAt should be set")
			}
			if cred.UpdatedAt.IsZero() {
				t.Fatal("UpdatedAt should be set")
			}
			credentials = append(credentials, cred)
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			want := credentials[0]
			opts := GetSiteCredentialOpts{ID: want.ID}

			have, err := s.GetSiteCredential(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("ByKind-URL", func(t *testing.T) {
			want := credentials[0]
			opts := GetSiteCredentialOpts{
				ExternalServiceType: want.ExternalServiceType,
				ExternalServiceID:   want.ExternalServiceID,
			}

			have, err := s.GetSiteCredential(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetSiteCredentialOpts{ID: 0xdeadbeef}

			_, have := s.GetSiteCredential(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			cs, next, err := s.ListSiteCredentials(ctx, ListSiteCredentialsOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if have, want := next, int64(0); have != want {
				t.Fatalf("have next %d, want %d", have, want)
			}

			have, want := cs, credentials
			if len(have) != len(want) {
				t.Fatalf("listed %d site credentials, want: %d", len(have), len(want))
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(credentials); i++ {
				cs, next, err := s.ListSiteCredentials(ctx, ListSiteCredentialsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(credentials) {
						want = credentials[i].ID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, credentials[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d site credentials, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			for _, cred := range credentials {
				if err := s.DeleteSiteCredential(ctx, cred.ID); err != nil {
					t.Fatal(err)
				}
			}
		})
		t.Run("NotFound", func(t *testing.T) {
			if err := s.DeleteSiteCredential(ctx, 0xdeadbeef); err == nil {
				t.Fatal("expected err but got nil")
			} else if err != ErrNoResults {
				t.Fatalf("invalid error %+v", err)
			}
		})
	})
}
