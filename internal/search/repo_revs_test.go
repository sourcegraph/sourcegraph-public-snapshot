package search

import (
	"reflect"
	"testing"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]struct {
		repo string
		revs []RevisionSpecifier
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
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repo, revs := ParseRepositoryRevisions(input)
			if repo != want.repo {
				t.Fatalf("got %+v, want %+v", repo, want.repo)
			}
			if !reflect.DeepEqual(revs, want.revs) {
				t.Fatalf("got %+v, want %+v", revs, want.revs)
			}
		})
	}
}
