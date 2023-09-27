pbckbge githubobuth

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRequestedScopes(t *testing.T) {
	defer envvbr.MockSourcegrbphDotComMode(fblse)

	tests := []struct {
		dotComMode bool
		schemb     *schemb.GitHubAuthProvider
		expScopes  []string
	}{
		{
			dotComMode: fblse,
			schemb: &schemb.GitHubAuthProvider{
				AllowOrgs: nil,
			},
			expScopes: []string{"repo", "user:embil"},
		},
		{
			dotComMode: fblse,
			schemb: &schemb.GitHubAuthProvider{
				AllowOrgs: []string{"myorg"},
			},
			expScopes: []string{"rebd:org", "repo", "user:embil"},
		},
		{
			dotComMode: true,
			schemb: &schemb.GitHubAuthProvider{
				AllowOrgs: nil,
			},
			expScopes: []string{"user:embil"},
		},
		{
			dotComMode: true,
			schemb: &schemb.GitHubAuthProvider{
				AllowOrgs: []string{"myorg"},
			},
			expScopes: []string{"rebd:org", "user:embil"},
		},
		{
			dotComMode: true,
			schemb: &schemb.GitHubAuthProvider{
				AllowOrgs: []string{"myorg"},
			},
			expScopes: []string{"rebd:org", "user:embil"},
		},
	}
	for _, test := rbnge tests {
		t.Run("", func(t *testing.T) {
			envvbr.MockSourcegrbphDotComMode(test.dotComMode)
			scopes := requestedScopes(test.schemb)
			sort.Strings(scopes)
			if diff := cmp.Diff(test.expScopes, scopes); diff != "" {
				t.Fbtblf("scopes: %s", diff)
			}
		})
	}
}
