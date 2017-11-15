package graphqlbackend

import (
	"context"
	"reflect"
	"sync"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestSearch2Suggestions(t *testing.T) {
	listOpts := sourcegraph.ListOptions{PerPage: int32(maxReposToSearch + 1)}

	createSearchResolver2 := func(t *testing.T, query, scopeQuery string) *searchResolver2 {
		args := &searchArgs2{Query: query, ScopeQuery: scopeQuery}
		r, err := (&rootResolver{}).Search2(args)
		if err != nil {
			t.Fatal("Search2:", err)
		}
		return r
	}
	getSuggestions := func(t *testing.T, query, scopeQuery string) []string {
		r := createSearchResolver2(t, query, scopeQuery)
		results, err := r.Suggestions(context.Background(), &searchSuggestionsArgs{})
		if err != nil {
			t.Fatal("Suggestions:", err)
		}
		resultDescriptions := make([]string, len(results))
		for i, result := range results {
			resultDescriptions[i] = testStringResult(result)
		}
		return resultDescriptions
	}
	testSuggestions := func(t *testing.T, query, scopeQuery string, want []string) {
		got := getSuggestions(t, query, scopeQuery)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got != want\ngot:  %v\nwant: %v", got, want)
		}
	}

	t.Run("empty", func(t *testing.T) {
		testSuggestions(t, "", "", []string{})
	})

	t.Run("whitespace", func(t *testing.T) {
		testSuggestions(t, " ", " ", []string{})
	})

	t.Run("single term", func(t *testing.T) {
		var calledReposListAll, calledReposListFoo bool
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			wantFoo := &store.RepoListOp{IncludePatterns: []string{"(^|/)foo"}, ListOptions: listOpts} // when treating term as repo: field
			wantAll := &store.RepoListOp{ListOptions: listOpts}                                        // when treating term as text query
			if reflect.DeepEqual(op, wantAll) {
				calledReposListAll = true
				return []*sourcegraph.Repo{{URI: "bar-repo"}}, nil
			} else if reflect.DeepEqual(op, wantFoo) {
				calledReposListFoo = true
				return []*sourcegraph.Repo{{URI: "foo-repo"}}, nil
			} else {
				t.Errorf("got %+v, want %+v or %+v", op, wantFoo, wantAll)
			}
			return nil, nil
		}
		store.Mocks.Repos.MockGetByURI(t, "repo", 1)
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) (*searchResults2, error) {
			calledSearchRepos = true
			if want := "foo"; args.Query.Pattern != want {
				t.Errorf("got %q, want %q", args.Query.Pattern, want)
			}
			return &searchResults2{
				results: []*fileMatch{
					{uri: "git://repo?rev#dir/file", JPath: "dir/file"},
				},
			}, nil
		}
		defer func() { mockSearchRepos = nil }()
		testSuggestions(t, "foo", "", []string{"repo:foo-repo", "file:dir/file"})
		if !calledReposListAll {
			t.Error("!calledReposListAll")
		}
		if !calledReposListFoo {
			t.Error("!calledReposListFoo")
		}
		if !calledSearchRepos {
			t.Error("!calledSearchRepos")
		}
	})

	t.Run("repogroup: and single term", func(t *testing.T) {
		var mu sync.Mutex
		var calledReposListReposInGroup, calledReposListFooRepo3 bool
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			wantReposInGroup := &store.RepoListOp{IncludePatterns: []string{`^foo-repo1$|^repo3$`}, ListOptions: listOpts}         // when treating term as repo: field
			wantFooRepo3 := &store.RepoListOp{IncludePatterns: []string{"(^|/)foo", `^foo-repo1$|^repo3$`}, ListOptions: listOpts} // when treating term as repo: field
			if reflect.DeepEqual(op, wantReposInGroup) {
				calledReposListReposInGroup = true
				return []*sourcegraph.Repo{{URI: "foo-repo1"}, {URI: "repo3"}}, nil
			} else if reflect.DeepEqual(op, wantFooRepo3) {
				calledReposListFooRepo3 = true
				return []*sourcegraph.Repo{{URI: "foo-repo1"}}, nil
			}
			t.Errorf("got %+v, want %+v or %+v", op, wantReposInGroup, wantFooRepo3)
			return nil, nil
		}
		store.Mocks.Repos.MockGetByURI(t, "repo", 1)
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) (*searchResults2, error) {
			mu.Lock()
			defer mu.Unlock()
			calledSearchRepos = true
			if want := "foo"; args.Query.Pattern != want {
				t.Errorf("got %q, want %q", args.Query.Pattern, want)
			}
			return &searchResults2{
				results: []*fileMatch{
					{uri: "git://repo?rev#dir/file-content-match", JPath: "dir/file-content-match"},
				},
			}, nil
		}
		var calledSearchFilesFoo, calledSearchFilesRepo3 bool
		defer func() { mockSearchRepos = nil }()
		mockSearchFilesForRepoURI = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
			mu.Lock()
			defer mu.Unlock()
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if repoURI == "foo-repo1" {
				calledSearchFilesFoo = true
				return []*searchResultResolver{
					{result: &fileResolver{path: "dir/foo-repo1-file-name-match", commit: commitSpec{RepoID: 1}}, score: 1},
				}, nil
			} else if repoURI == "repo3" {
				calledSearchFilesRepo3 = true
				return []*searchResultResolver{
					{result: &fileResolver{path: "dir/foo-repo3-file-name-match", commit: commitSpec{RepoID: 1}}, score: 2},
				}, nil
			}
			t.Errorf("got %q, want %q or %q", repoURI, "foo-repo1", "repo3")
			return nil, nil
		}
		defer func() { mockSearchFilesForRepoURI = nil }()
		calledResolveRepoGroups := false
		defer func() { mockResolveRepoGroups = nil }()
		mockResolveRepoGroups = func() (map[string][]*sourcegraph.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledResolveRepoGroups = true
			return map[string][]*sourcegraph.Repo{
				"sample": []*sourcegraph.Repo{{URI: "foo-repo1"}, {URI: "repo3"}},
			}, nil
		}
		testSuggestions(t, "foo", "repogroup:sample", []string{"repo:foo-repo1", "file:dir/foo-repo3-file-name-match", "file:dir/foo-repo1-file-name-match", "file:dir/file-content-match"})
		if !calledReposListReposInGroup {
			t.Error("!calledReposListReposInGroup")
		}
		if !calledReposListFooRepo3 {
			t.Error("!calledReposListFooRepo3")
		}
		if !calledSearchRepos {
			t.Error("!calledSearchRepos")
		}
		if !calledSearchFilesFoo {
			t.Error("!calledSearchFilesFoo")
		}
		if !calledSearchFilesRepo3 {
			t.Error("!calledSearchFilesRepo3")
		}
		if !calledResolveRepoGroups {
			t.Error("!calledResolveRepoGroups")
		}
	})

	t.Run("repo: field", func(t *testing.T) {
		var mu sync.Mutex
		calledReposList := false
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (&store.RepoListOp{IncludePatterns: []string{"(^|/)foo"}, ListOptions: listOpts}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*sourcegraph.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepoURI = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := "foo-repo"; repoURI != want {
				t.Errorf("got %q, want %q", repoURI, want)
			}
			return []*searchResultResolver{
				{result: &fileResolver{path: "dir/file", commit: commitSpec{RepoID: 1}}, score: 1},
			}, nil
		}
		defer func() { mockSearchFilesForRepoURI = nil }()
		testSuggestions(t, "repo:foo", "", []string{"repo:foo-repo", "file:dir/file"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchFiles {
			t.Error("!calledSearchFiles")
		}
	})

	t.Run("repo: and file: field", func(t *testing.T) {
		var mu sync.Mutex
		calledReposList := false
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (&store.RepoListOp{IncludePatterns: []string{"(^|/)foo"}, ListOptions: listOpts}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*sourcegraph.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepoURI = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := "foo-repo"; repoURI != want {
				t.Errorf("got %q, want %q", repoURI, want)
			}
			return []*searchResultResolver{
				{result: &fileResolver{path: "dir/bar-file", commit: commitSpec{RepoID: 1}}, score: 1},
			}, nil
		}
		defer func() { mockSearchFilesForRepoURI = nil }()
		testSuggestions(t, "repo:foo file:bar", "", []string{"file:dir/bar-file"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchFiles {
			t.Error("!calledSearchFiles")
		}
	})
}
