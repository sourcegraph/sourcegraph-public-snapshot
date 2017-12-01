package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParseRepositoryRevisions(t *testing.T) {
	tests := map[string]repositoryRevisions{
		"repo":     repositoryRevisions{repo: "repo", revspecs: nil},
		"repo@rev": repositoryRevisions{repo: "repo", revspecs: []string{"rev"}},
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			repoRevs := parseRepositoryRevisions(input)
			if !reflect.DeepEqual(repoRevs, want) {
				t.Fatalf("got %+v, want %+v", repoRevs, want)
			}
		})
	}
}
