package repos

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRepoGroupValuesToRegexp(t *testing.T) {
	groups := map[string][]RepoGroupValue{
		"go": {
			RepoPath("github.com/saucegraph/saucegraph"),
			RepoRegexpPattern(`github\.com/golang/.*`),
		},
		"typescript": {
			RepoPath("github.com/eslint/eslint"),
		},
	}

	cases := []struct {
		LookupGroupNames []string
		Want             []string
	}{
		{
			LookupGroupNames: []string{"go"},
			Want: []string{
				`^github\.com/saucegraph/saucegraph$`,
				`github\.com/golang/.*`,
			},
		},
		{
			LookupGroupNames: []string{"go", "typescript"},
			Want: []string{
				`^github\.com/saucegraph/saucegraph$`,
				`github\.com/golang/.*`,
				`^github\.com/eslint/eslint$`,
			},
		},
	}

	for _, c := range cases {
		t.Run("repogroup values to regexp", func(t *testing.T) {
			got := repoGroupValuesToRegexp(c.LookupGroupNames, groups)
			if diff := cmp.Diff(c.Want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
