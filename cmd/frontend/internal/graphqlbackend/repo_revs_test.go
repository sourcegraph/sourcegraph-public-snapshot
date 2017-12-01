package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParseRepositoryRevisions(t *testing.T) {
	wantString := map[string]string{"repo@": "repo", "repo@:rev1:": "repo@rev1"}
	tests := map[string]repositoryRevisions{
		"repo":           repositoryRevisions{repo: "repo"},
		"repo@":          repositoryRevisions{repo: "repo"},
		"repo@rev":       repositoryRevisions{repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev"}}},
		"repo@rev1:rev2": repositoryRevisions{repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev1"}, {revspec: "rev2"}}},
		"repo@:rev1:":    repositoryRevisions{repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev1"}}},
		"repo@*glob":     repositoryRevisions{repo: "repo", revs: []revspecOrRefGlob{{refGlob: "glob"}}},
		"repo@rev1:*glob1:^rev2": repositoryRevisions{
			repo: "repo",
			revs: []revspecOrRefGlob{{revspec: "rev1"}, {refGlob: "glob1"}, {revspec: "^rev2"}},
		},
		"repo@rev1:*glob1:*!glob2:rev2:*glob3": repositoryRevisions{
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
			repoRevs := parseRepositoryRevisions(input)
			if !reflect.DeepEqual(repoRevs, want) {
				t.Fatalf("got %+v, want %+v", repoRevs, want)
			}

			wantStr := wantString[input]
			if wantStr == "" {
				wantStr = input
			}
			if str := repoRevs.String(); str != wantStr {
				t.Fatalf("got %q, want %q", str, wantStr)
			}
		})
	}
}
