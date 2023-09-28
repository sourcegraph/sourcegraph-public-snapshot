package types

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestBatchChange_URL(t *testing.T) {
	ctx := context.Background()
	bc := &BatchChange{Name: "bar", NamespaceOrgID: 123}

	t.Run("errors", func(t *testing.T) {
		for name, url := range map[string]string{
			"invalid URL": "foo://:bar",
		} {
			t.Run(name, func(t *testing.T) {
				mockExternalURL(t, url)
				if _, err := bc.URL(ctx, "namespace"); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		mockExternalURL(t, "https://sourcegraph.test")
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

func mockExternalURL(t *testing.T, url string) {
	oldConf := conf.Get()
	newConf := *oldConf
	newConf.ExternalURL = url
	conf.Mock(&newConf)
	t.Cleanup(func() { conf.Mock(oldConf) })
}
