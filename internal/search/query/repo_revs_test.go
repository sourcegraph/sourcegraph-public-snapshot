package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]ParsedRepoFilter{
		"repo":           {Repo: "repo", Revs: []RevisionSpecifier{}},
		"repo@":          {Repo: "repo", Revs: []RevisionSpecifier{{RevSpec: ""}}},
		"repo@rev":       {Repo: "repo", Revs: []RevisionSpecifier{{RevSpec: "rev"}}},
		"repo@rev1:rev2": {Repo: "repo", Revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RevSpec: "rev2"}}},
		"repo@:rev1:":    {Repo: "repo", Revs: []RevisionSpecifier{{RevSpec: "rev1"}}},
		"repo@*glob":     {Repo: "repo", Revs: []RevisionSpecifier{{RefGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": {
			Repo: "repo",
			Revs: []RevisionSpecifier{{RevSpec: "rev1"}, {RefGlob: "glob1"}, {RevSpec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": {
			Repo: "repo",
			Revs: []RevisionSpecifier{
				{RevSpec: "rev1"},
				{RefGlob: "glob1"},
				{ExcludeRefGlob: "glob2"},
				{RevSpec: "rev2"},
				{RefGlob: "glob3"},
			},
		},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repoRevs := ParseRepositoryRevisions(input)
			if diff := cmp.Diff(want, repoRevs); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}
