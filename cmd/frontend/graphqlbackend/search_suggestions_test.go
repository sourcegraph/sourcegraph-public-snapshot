package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSearchSuggestions(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch() + 1}

	createSearchResolver := func(t *testing.T, query string) *searchResolver {
		t.Helper()
		r, err := (&schemaResolver{}).Search(&struct{ Query string }{Query: query})
		if err != nil {
			t.Fatal("Search:", err)
		}
		return r
	}
	getSuggestions := func(t *testing.T, query string) []string {
		t.Helper()
		r := createSearchResolver(t, query)
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
	testSuggestions := func(t *testing.T, query string, want []string) {
		t.Helper()
		got := getSuggestions(t, query)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got != want\ngot:  %v\nwant: %v", got, want)
		}
	}

	t.Run("empty", func(t *testing.T) {
		testSuggestions(t, "", []string{})
	})

	t.Run("whitespace", func(t *testing.T) {
		testSuggestions(t, " ", []string{})
	})

	t.Run("single term", func(t *testing.T) {
		var calledReposListAll, calledReposListFoo bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			wantFoo := db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset} // when treating term as repo: field
			wantAll := db.ReposListOptions{Enabled: true, LimitOffset: limitOffset}                                   // when treating term as text query
			if reflect.DeepEqual(op, wantAll) {
				calledReposListAll = true
				return []*types.Repo{{URI: "bar-repo"}}, nil
			} else if reflect.DeepEqual(op, wantFoo) {
				calledReposListFoo = true
				return []*types.Repo{{URI: "foo-repo"}}, nil
			} else {
				t.Errorf("got %+v, want %+v or %+v", op, wantFoo, wantAll)
			}
			return nil, nil
		}
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)
		backend.Mocks.Repos.MockResolveRev_NoCheck(t, api.CommitID("deadbeef"))
		defer func() { db.Mocks = db.MockStores{} }()

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *search.Args) ([]*fileMatchResolver, *searchResultsCommon, error) {
			calledSearchFilesInRepos = true
			if want := "foo"; args.Pattern.Pattern != want {
				t.Errorf("got %q, want %q", args.Pattern.Pattern, want)
			}
			return []*fileMatchResolver{
				{uri: "git://repo?rev#dir/file", JPath: "dir/file", repo: &types.Repo{URI: "repo"}, commitID: "rev"},
			}, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		testSuggestions(t, "foo", []string{"repo:foo-repo", "file:dir/file"})
		if !calledReposListAll {
			t.Error("!calledReposListAll")
		}
		if !calledReposListFoo {
			t.Error("!calledReposListFoo")
		}
		if !calledSearchFilesInRepos {
			t.Error("!calledSearchFilesInRepos")
		}
	})

	t.Run("single term invalid regex", func(t *testing.T) {
		_, err := (&schemaResolver{}).Search(&struct{ Query string }{Query: "foo("})
		if err == nil {
			t.Fatal("err == nil")
		} else if want := "error parsing regexp"; !strings.Contains(err.Error(), want) {
			t.Fatalf("got error %q, want it to contain %q", err, want)
		}
	})

	t.Run("repogroup: and single term", func(t *testing.T) {
		var mu sync.Mutex
		var calledReposListReposInGroup, calledReposListFooRepo3 bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			wantReposInGroup := db.ReposListOptions{IncludePatterns: []string{`^foo-repo1$|^repo3$`}, Enabled: true, LimitOffset: limitOffset}    // when treating term as repo: field
			wantFooRepo3 := db.ReposListOptions{IncludePatterns: []string{"foo", `^foo-repo1$|^repo3$`}, Enabled: true, LimitOffset: limitOffset} // when treating term as repo: field
			if reflect.DeepEqual(op, wantReposInGroup) {
				calledReposListReposInGroup = true
				return []*types.Repo{{URI: "foo-repo1"}, {URI: "repo3"}}, nil
			} else if reflect.DeepEqual(op, wantFooRepo3) {
				calledReposListFooRepo3 = true
				return []*types.Repo{{URI: "foo-repo1"}}, nil
			}
			t.Errorf("got %+v, want %+v or %+v", op, wantReposInGroup, wantFooRepo3)
			return nil, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)
		backend.Mocks.Repos.MockResolveRev_NoCheck(t, api.CommitID("deadbeef"))

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *search.Args) ([]*fileMatchResolver, *searchResultsCommon, error) {
			mu.Lock()
			defer mu.Unlock()
			calledSearchFilesInRepos = true
			if want := "foo"; args.Pattern.Pattern != want {
				t.Errorf("got %q, want %q", args.Pattern.Pattern, want)
			}
			return []*fileMatchResolver{
				{uri: "git://repo?rev#dir/file-content-match", JPath: "dir/file-content-match", repo: &types.Repo{URI: "repo"}, commitID: "rev"},
			}, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		var calledSearchFilesFoo, calledSearchFilesRepo3 bool
		mockSearchFilesForRepo = func(matcher matcher, repoRevs search.RepositoryRevisions, limit int, includeDirs bool) ([]*searchSuggestionResolver, error) {
			mu.Lock()
			defer mu.Unlock()
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if repoRevs.Repo.URI == "foo-repo1" {
				calledSearchFilesFoo = true
				return []*searchSuggestionResolver{
					{result: &gitTreeEntryResolver{path: "dir/foo-repo1-file-name-match", commit: &gitCommitResolver{repo: &repositoryResolver{repo: &types.Repo{URI: "r"}}}}, score: 1},
				}, nil
			} else if repoRevs.Repo.URI == "repo3" {
				calledSearchFilesRepo3 = true
				return []*searchSuggestionResolver{
					{result: &gitTreeEntryResolver{path: "dir/foo-repo3-file-name-match", commit: &gitCommitResolver{repo: &repositoryResolver{repo: &types.Repo{URI: "r"}}}}, score: 2},
				}, nil
			}
			t.Errorf("got %q, want %q or %q", repoRevs.Repo.URI, "foo-repo1", "repo3")
			return nil, nil
		}
		defer func() { mockSearchFilesForRepo = nil }()

		calledResolveRepoGroups := false
		mockResolveRepoGroups = func() (map[string][]*types.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledResolveRepoGroups = true
			return map[string][]*types.Repo{
				"sample": {{URI: "foo-repo1"}, {URI: "repo3"}},
			}, nil
		}
		defer func() { mockResolveRepoGroups = nil }()

		mockSearchSymbols = func(ctx context.Context, args *search.Args, limit int) (res []*fileMatchResolver, common *searchResultsCommon, err error) {
			// TODO test symbol suggestions
			return nil, nil, nil
		}
		defer func() { mockSearchSymbols = nil }()

		testSuggestions(t, "repogroup:sample foo", []string{"repo:foo-repo1", "file:dir/foo-repo3-file-name-match", "file:dir/foo-repo1-file-name-match", "file:dir/file-content-match"})
		if !calledReposListReposInGroup {
			t.Error("!calledReposListReposInGroup")
		}
		if !calledReposListFooRepo3 {
			t.Error("!calledReposListFooRepo3")
		}
		if !calledSearchFilesInRepos {
			t.Error("!calledSearchFilesInRepos")
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*types.Repo{{URI: "foo-repo"}}, nil
		}
		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoRevs search.RepositoryRevisions, limit int, includeDirs bool) ([]*searchSuggestionResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := api.RepoURI("foo-repo"); repoRevs.Repo.URI != want {
				t.Errorf("got %q, want %q", repoRevs.Repo.URI, want)
			}
			return []*searchSuggestionResolver{
				{result: &gitTreeEntryResolver{path: "dir/file", commit: &gitCommitResolver{repo: &repositoryResolver{repo: &types.Repo{URI: "r"}}}}, score: 1},
			}, nil
		}
		defer func() { mockSearchFilesForRepo = nil }()
		testSuggestions(t, "repo:foo", []string{"repo:foo-repo", "file:dir/file"})
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
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposList = true
			if want := (db.ReposListOptions{IncludePatterns: []string{"foo"}, Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Errorf("got %+v, want %+v", op, want)
			}
			return []*types.Repo{{URI: "foo-repo"}}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()

		calledSearchFiles := false
		mockSearchFilesForRepo = func(matcher matcher, repoRevs search.RepositoryRevisions, limit int, includeDirs bool) ([]*searchSuggestionResolver, error) {
			calledSearchFiles = true
			if want := ""; matcher.query != want {
				t.Errorf("got %q, want %q", matcher.query, want)
			}
			if want := api.RepoURI("foo-repo"); repoRevs.Repo.URI != want {
				t.Errorf("got %q, want %q", repoRevs.Repo.URI, want)
			}
			return []*searchSuggestionResolver{
				{result: &gitTreeEntryResolver{path: "dir/bar-file", commit: &gitCommitResolver{repo: &repositoryResolver{repo: &types.Repo{URI: "r"}}}}, score: 1},
			}, nil
		}
		defer func() { mockSearchFilesForRepo = nil }()

		testSuggestions(t, "repo:foo file:bar", []string{"file:dir/bar-file"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchFiles {
			t.Error("!calledSearchFiles")
		}
	})
}
