package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestSearchSuggestions(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch + 1}

	createSearchResolver := func(t *testing.T, query, scopeQuery string) *searchResolver {
		args := &searchArgs{Query: query, ScopeQuery: scopeQuery}
		r, err := (&schemaResolver{}).Search(args)
		if err != nil {
			t.Fatal("Search:", err)
		}
		return r
	}
	getSuggestions := func(t *testing.T, query, scopeQuery string) []string {
		r := createSearchResolver(t, query, scopeQuery)
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*api.Repo, error) {
			wantFoo := db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset} // when treating term as repo: field
			wantAll := db.ReposListOptions{Enabled: true, LimitOffset: limitOffset}                                   // when treating term as text query
			if reflect.DeepEqual(op, wantAll) {
				calledReposListAll = true
				return []*api.Repo{{URI: "bar-repo"}}, nil
			} else if reflect.DeepEqual(op, wantFoo) {
				calledReposListFoo = true
				return []*api.Repo{{URI: "foo-repo"}}, nil
			} else {
				t.Errorf("got %+v, want %+v or %+v", op, wantFoo, wantAll)
			}
			return nil, nil
		}
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)
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
		args := &searchArgs{Query: "foo(", ScopeQuery: ""}
		_, err := (&schemaResolver{}).Search(args)
		if err == nil {
			t.Fatal("err == nil")
		} else if want := "error parsing regexp"; !strings.Contains(err.Error(), want) {
			t.Fatalf("got error %q, want it to contain %q", err, want)
		}
	})

	t.Run("repogroup: and single term", func(t *testing.T) {
		var mu sync.Mutex
		var calledReposListReposInGroup, calledReposListFooRepo3 bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*api.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			wantReposInGroup := db.ReposListOptions{IncludePatterns: []string{`^foo-repo1$|^repo3$`}, Enabled: true, LimitOffset: limitOffset}    // when treating term as repo: field
			wantFooRepo3 := db.ReposListOptions{IncludePatterns: []string{"foo", `^foo-repo1$|^repo3$`}, Enabled: true, LimitOffset: limitOffset} // when treating term as repo: field
			if reflect.DeepEqual(op, wantReposInGroup) {
				calledReposListReposInGroup = true
				return []*api.Repo{{URI: "foo-repo1"}, {URI: "repo3"}}, nil
			} else if reflect.DeepEqual(op, wantFooRepo3) {
				calledReposListFooRepo3 = true
				return []*api.Repo{{URI: "foo-repo1"}}, nil
			}
			t.Errorf("got %+v, want %+v or %+v", op, wantReposInGroup, wantFooRepo3)
			return nil, nil
		}
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)
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
		mockSearchFilesForRepo = func(matcher matcher, repoRevs repositoryRevisions, limit int, includeDirs bool) ([]*searchResultResolver, error) {
			mu.Lock()
			defer mu.Unlock()
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if repoRevs.repo.URI == "foo-repo1" {
				calledSearchFilesFoo = true
				return []*searchResultResolver{
					{result: &fileResolver{path: "dir/foo-repo1-file-name-match", commit: &gitCommitResolver{repoID: 1}}, score: 1},
				}, nil
			} else if repoRevs.repo.URI == "repo3" {
				calledSearchFilesRepo3 = true
				return []*searchResultResolver{
					{result: &fileResolver{path: "dir/foo-repo3-file-name-match", commit: &gitCommitResolver{repoID: 1}}, score: 2},
				}, nil
			}
			t.Errorf("got %q, want %q or %q", repoRevs.repo.URI, "foo-repo1", "repo3")
			return nil, nil
		}
		defer func() { mockSearchFilesForRepo = nil }()
		calledResolveRepoGroups := false
		defer func() { mockResolveRepoGroups = nil }()
		mockResolveRepoGroups = func() (map[string][]*api.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledResolveRepoGroups = true
			return map[string][]*api.Repo{
				"sample": []*api.Repo{{URI: "foo-repo1"}, {URI: "repo3"}},
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*api.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*api.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoRevs repositoryRevisions, limit int, includeDirs bool) ([]*searchResultResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := "foo-repo"; repoRevs.repo.URI != want {
				t.Errorf("got %q, want %q", repoRevs.repo.URI, want)
			}
			return []*searchResultResolver{
				{result: &fileResolver{path: "dir/file", commit: &gitCommitResolver{repoID: 1}}, score: 1},
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*api.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*api.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoRevs repositoryRevisions, limit int, includeDirs bool) ([]*searchResultResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := "foo-repo"; repoRevs.repo.URI != want {
				t.Errorf("got %q, want %q", repoRevs.repo.URI, want)
			}
			return []*searchResultResolver{
				{result: &fileResolver{path: "dir/bar-file", commit: &gitCommitResolver{repoID: 1}}, score: 1},
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
