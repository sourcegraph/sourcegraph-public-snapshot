package types

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBatchChange_URL(t *testing.T) {
	ctx := context.Background()
	bc := &BatchChange{Name: "bar", NamespaceOrgID: 123}

	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]*mockInternalClient{
			"ExternalURL error": {err: errors.New("foo")},
			"invalid URL":       {externalURL: "foo://:bar"},
		} {
			t.Run(name, func(t *testing.T) {
				MockInternalClient(tc)
				t.Cleanup(ResetInternalClient)

				if _, err := bc.URL(ctx, "namespace"); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		MockInternalClientExternalURL("https://sourcegraph.test")
		t.Cleanup(ResetInternalClient)

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
