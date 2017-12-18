package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]struct {
		repo string
		revs []revspecOrRefGlob
	}{
		"repo":           {repo: "repo"},
		"repo@":          {repo: "repo"},
		"repo@rev":       {repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev"}}},
		"repo@rev1:rev2": {repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev1"}, {revspec: "rev2"}}},
		"repo@:rev1:":    {repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev1"}}},
		"repo@*glob":     {repo: "repo", revs: []revspecOrRefGlob{{refGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": {
			repo: "repo",
			revs: []revspecOrRefGlob{{revspec: "rev1"}, {refGlob: "glob1"}, {revspec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": {
			repo: "repo",
			revs: []revspecOrRefGlob{
				{revspec: "rev1"},
				{refGlob: "glob1"},
				{excludeRefGlob: "glob2"},
				{revspec: "rev2"},
				{refGlob: "glob3"},
			},
		},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repo, revs := parseRepositoryRevisions(input)
			if repo != want.repo {
				t.Fatalf("got %+v, want %+v", repo, want.repo)
			}
			if !reflect.DeepEqual(revs, want.revs) {
				t.Fatalf("got %+v, want %+v", revs, want.revs)
			}
		})
	}
}
