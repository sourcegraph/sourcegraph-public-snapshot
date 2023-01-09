package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp/syntax"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]struct {
		repo string
		revs []RevisionSpecifier
		err  error
	}{
		"repo":           {repo: "repo", revs: []RevisionSpecifier{}},
		"repo@":          {repo: "repo", revs: []RevisionSpecifier{{RevSpec: ""}}},
		"repo@rev":       {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev"}}},
		"repo@rev1:rev2": {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RevSpec: "rev2"}}},
		"repo@:rev1:":    {repo: "repo", revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo@*glob":     {repo: "repo", revs: []RevisionSpecifier{{RefGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": {
			repo: "repo",
			revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RefGlob: "glob1"}, {RevSpec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": {
			repo: "repo",
			revs: []RevisionSpecifier{
				{RevSpec: "rev1"},
				{RefGlob: "glob1"},
				{ExcludeRefGlob: "glob2"},
				{RevSpec: "rev2"},
				{RefGlob: "glob3"},
			},
		},
		"@rev1":            {repo: "", revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo?*@rev1:rev2": {err: &syntax.Error{Code: "invalid nested repetition operator", Expr: "?*"}},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repoRevs, err := ParseRepositoryRevisions(input)
			if diff := cmp.Diff(errors.UnwrapAll(want.err), err); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(want.repo, repoRevs.Repo); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}

			// Just check the repo regex is present -- there are other tests
			// that exercise the compiled regex
			if err == nil && repoRevs.RepoRegex == nil {
				t.Fatalf("repo regex is unexpectedly empty")
			}

			if diff := cmp.Diff(want.revs, repoRevs.Revs); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}
