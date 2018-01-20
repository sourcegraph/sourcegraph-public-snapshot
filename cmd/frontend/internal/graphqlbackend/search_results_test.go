package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSearchResults(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch + 1}

	createSearchResolver := func(t *testing.T, query string) *searchResolver {
		args := &searchArgs{Query: query}
		r, err := (&schemaResolver{}).Search(args)
		if err != nil {
			t.Fatal("Search:", err)
		}
		return r
	}
	getResults := func(t *testing.T, query string) []string {
		r := createSearchResolver(t, query)
		results, err := r.Results(context.Background())
		if err != nil {
			t.Fatal("Results:", err)
		}
		resultDescriptions := make([]string, len(results.results))
		for i, result := range results.results {
			// NOTE: Only supports one match per line. If we need to test other cases,
			// just remove that assumption in the following line of code.
			resultDescriptions[i] = fmt.Sprintf("%s:%d", result.fileMatch.JPath, result.fileMatch.JLineMatches[0].JLineNumber)
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*api.Repo, error) {
			calledReposList = true
			if want := (db.ReposListOptions{Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*api.Repo{{URI: "repo"}}, nil
		}
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error) {
			calledSearchRepos = true
			if want := `(foo\d).*?(bar\*)`; args.query.Pattern != want {
				t.Errorf("got %q, want %q", args.query.Pattern, want)
			}
			return fileMatchesToSearchResults([]*fileMatch{
				{uri: "git://repo?rev#dir/file", JPath: "dir/file", JLineMatches: []*lineMatch{{JLineNumber: 123}}},
			}), &searchResultsCommon{}, nil
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

func TestRegexpPatternMatchingExprsInOrder(t *testing.T) {
	got := regexpPatternMatchingExprsInOrder([]string{})
	if want := ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = regexpPatternMatchingExprsInOrder([]string{"a"})
	if want := "a"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = regexpPatternMatchingExprsInOrder([]string{"a", "b|c"})
	if want := "(a).*?(b|c)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
