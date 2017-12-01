package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]repositoryRevisions{
		"repo":           repositoryRevisions{repo: "repo", revspecs: nil},
		"repo@rev":       repositoryRevisions{repo: "repo", revspecs: []string{"rev"}},
		"repo@rev1:rev2": repositoryRevisions{repo: "repo", revspecs: []string{"rev1", "rev2"}},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repoRevs := parseRepositoryRevisions(input)
			if !reflect.DeepEqual(repoRevs, want) {
				t.Fatalf("got %+v, want %+v", repoRevs, want)
			}
			if str := repoRevs.String(); str != input {
				t.Fatalf("got %q, want %q", str, input)
			}
		})
	}
}
