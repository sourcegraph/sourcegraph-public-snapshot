package types

import (
	"context"
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestBatchChange_URL(t *testing.T) {
	ctx := context.Background()
	bc := &BatchChange{Name: "bar", NamespaceOrgID: 123}
	globals.SetExternalURL(&url.URL{Scheme: "foo", Host: "//:bar"})
	t.Run("errors", func(t *testing.T) {
		for _, name := range []string{
			"invalid URL",
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := bc.URL(ctx, "namespace"); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		globals.SetExternalURL(&url.URL{Scheme: "https", Host: "sourcegraph.test"})
		url, err := bc.URL(
			ctx,
			"foo",
		)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if want := "https://sourcegraph.test/organizations/foo/batch-changes/bar"; url != want {
			t.Errorf("unexpected URL: have=%q want=%q", url, want)
		}
	})
}

func TestNamespaceURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		ns   *database.Namespace
		want string
	}{
		"user": {
			ns:   &database.Namespace{User: 123, Name: "user"},
			want: "/users/user",
		},
		"org": {
			ns:   &database.Namespace{Organization: 123, Name: "org"},
			want: "/organizations/org",
		},
		"neither": {
			ns:   &database.Namespace{Name: "user"},
			want: "/users/user",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := namespaceURL(tc.ns.Organization, tc.ns.Name); have != tc.want {
				t.Errorf("unexpected URL: have=%q want=%q", have, tc.want)
			}
		})
	}
}
