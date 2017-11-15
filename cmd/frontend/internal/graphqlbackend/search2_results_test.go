package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestSearch2Results(t *testing.T) {
	listOpts := sourcegraph.ListOptions{PerPage: int32(maxReposToSearch + 1)}

	createSearchResolver2 := func(t *testing.T, query, scopeQuery string) *searchResolver2 {
		args := &searchArgs2{Query: query, ScopeQuery: scopeQuery}
		r, err := (&rootResolver{}).Search2(args)
		if err != nil {
			t.Fatal("Search2:", err)
		}
		return r
	}
	getResults := func(t *testing.T, query string) []string {
		r := createSearchResolver2(t, query, "")
		results, err := r.Results(context.Background())
		if err != nil {
			t.Fatal("Results:", err)
		}
		resultDescriptions := make([]string, len(results.results))
		for i, result := range results.results {
			// NOTE: Only supports one match per line. If we need to test other cases,
			// just remove that assumption in the following line of code.
			resultDescriptions[i] = fmt.Sprintf("%s:%d", result.JPath, result.JLineMatches[0].JLineNumber)
		}
		return resultDescriptions
	}
	testCallResults := func(t *testing.T, query string, want []string) {
		results := getResults(t, query)
		if !reflect.DeepEqual(results, want) {
			t.Errorf("got %v, want %v", results, want)
		}
	}

	t.Run("multiple terms", func(t *testing.T) {
		var calledReposList bool
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			calledReposList = true
			if want := (&store.RepoListOp{ListOptions: listOpts}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*sourcegraph.Repo{{URI: "repo"}}, nil
		}
		store.Mocks.Repos.MockGetByURI(t, "repo", 1)
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) (*searchResults2, error) {
			calledSearchRepos = true
			if want := `foo\d.*?bar\*`; args.Query.Pattern != want {
				t.Errorf("got %q, want %q", args.Query.Pattern, want)
			}
			return &searchResults2{
				results: []*fileMatch{
					{uri: "git://repo?rev#dir/file", JPath: "dir/file", JLineMatches: []*lineMatch{{JLineNumber: 123}}},
				},
			}, nil
		}
		defer func() { mockSearchRepos = nil }()
		testCallResults(t, `foo\d "bar*"`, []string{"dir/file:123"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchRepos {
			t.Error("!calledSearchRepos")
		}
	})
}
