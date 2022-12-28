package bitbucketcloudoauth

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRequestedScopes(t *testing.T) {
	defer envvar.MockSourcegraphDotComMode(false)

	tests := []struct {
		dotComMode bool
		schema     *schema.BitbucketCloudAuthProvider
		expScopes  []string
	}{
		{
			dotComMode: false,
			schema:     &schema.BitbucketCloudAuthProvider{},
			expScopes:  []string{"email", "repository"},
		},
		{
			dotComMode: true,
			schema:     &schema.BitbucketCloudAuthProvider{},
			expScopes:  []string{"email"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			envvar.MockSourcegraphDotComMode(test.dotComMode)
			scopes := requestedScopes()
			sort.Strings(scopes)
			if diff := cmp.Diff(test.expScopes, scopes); diff != "" {
				t.Fatalf("scopes: %s", diff)
			}
		})
	}
}
