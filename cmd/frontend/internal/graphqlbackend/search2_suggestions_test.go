package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
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
			wantFoo := &store.RepoListOp{IncludePatterns: []string{"foo"}, ListOptions: listOpts} // when treating term as repo: field
			wantAll := &store.RepoListOp{ListOptions: listOpts}                                   // when treating term as text query
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
		backend.Mocks.Repos.MockResolveRev_NoCheck(t, vcs.CommitID("deadbeef"))
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error) {
			calledSearchRepos = true
			if want := "foo"; args.query.Pattern != want {
				t.Errorf("got %q, want %q", args.query.Pattern, want)
			}
			return fileMatchesToSearchResults([]*fileMatch{
				{uri: "git://repo?rev#dir/file", JPath: "dir/file"},
			}), &searchResultsCommon{}, nil
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

	t.Run("single term invalid regex", func(t *testing.T) {
		args := &searchArgs2{Query: "foo(", ScopeQuery: ""}
		_, err := (&rootResolver{}).Search2(args)
		if err == nil {
			t.Fatal("err == nil")
		} else if want := "error parsing regexp"; !strings.Contains(err.Error(), want) {
			t.Fatalf("got error %q, want it to contain %q", err, want)
		}
	})

	t.Run("repogroup: and single term", func(t *testing.T) {
		var mu sync.Mutex
		var calledReposListReposInGroup, calledReposListFooRepo3 bool
		store.Mocks.Repos.List = func(_ context.Context, op *store.RepoListOp) ([]*sourcegraph.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			wantReposInGroup := &store.RepoListOp{IncludePatterns: []string{`^foo-repo1$|^repo3$`}, ListOptions: listOpts}    // when treating term as repo: field
			wantFooRepo3 := &store.RepoListOp{IncludePatterns: []string{"foo", `^foo-repo1$|^repo3$`}, ListOptions: listOpts} // when treating term as repo: field
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
		backend.Mocks.Repos.MockResolveRev_NoCheck(t, vcs.CommitID("deadbeef"))
		calledSearchRepos := false
		mockSearchRepos = func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error) {
			mu.Lock()
			defer mu.Unlock()
			calledSearchRepos = true
			if want := "foo"; args.query.Pattern != want {
				t.Errorf("got %q, want %q", args.query.Pattern, want)
			}
			return fileMatchesToSearchResults([]*fileMatch{
				{uri: "git://repo?rev#dir/file-content-match", JPath: "dir/file-content-match"},
			}), &searchResultsCommon{}, nil
		}
		var calledSearchFilesFoo, calledSearchFilesRepo3 bool
		defer func() { mockSearchRepos = nil }()
		mockSearchFilesForRepo = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
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
		defer func() { mockSearchFilesForRepo = nil }()
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
			if want := (&store.RepoListOp{IncludePatterns: []string{"foo"}, ListOptions: listOpts}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*sourcegraph.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
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
		defer func() { mockSearchFilesForRepo = nil }()
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
			if want := (&store.RepoListOp{IncludePatterns: []string{"foo"}, ListOptions: listOpts}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*sourcegraph.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoURI string, limit int) ([]*searchResultResolver, error) {
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
		defer func() { mockSearchFilesForRepo = nil }()
		testSuggestions(t, "repo:foo file:bar", "", []string{"file:dir/bar-file"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchFiles {
			t.Error("!calledSearchFiles")
		}
	})
}
