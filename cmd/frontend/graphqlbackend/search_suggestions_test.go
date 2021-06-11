package graphqlbackend

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchSuggestions(t *testing.T) {
	db := new(dbtesting.MockDB)

	getSuggestions := func(t *testing.T, query, version string) []string {
		t.Helper()
		r, err := (&schemaResolver{db: db}).Search(context.Background(), &SearchArgs{Query: query, Version: version})
		if err != nil {
			t.Fatal("Search:", err)
		}
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
	testSuggestions := func(t *testing.T, query, version string, want []string) {
		t.Helper()
		got := getSuggestions(t, query, version)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got != want\ngot:  %v\nwant: %v", got, want)
		}
	}

	run.MockSearchSymbols = func(ctx context.Context, args *search.TextParameters, limit int) (res []*result.FileMatch, common *streaming.Stats, err error) {
		// TODO test symbol suggestions
		return nil, nil, nil
	}
	defer func() { run.MockSearchSymbols = nil }()

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	searchVersions := []string{"V1", "V2"}

	t.Run("empty", func(t *testing.T) {
		for _, v := range searchVersions {
			testSuggestions(t, "", v, []string{})
		}
	})

	t.Run("whitespace", func(t *testing.T) {
		for _, v := range searchVersions {
			testSuggestions(t, " ", v, []string{})
		}
	})

	t.Run("single term", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		var calledReposListNamesAll, calledReposListFoo bool
		database.Mocks.Repos.ListRepoNames = func(_ context.Context, op database.ReposListOptions) ([]types.RepoName, error) {

			if reflect.DeepEqual(op.IncludePatterns, []string{"foo"}) {
				// when treating term as repo: field
				calledReposListFoo = true
				return []types.RepoName{{Name: "foo-repo"}}, nil
			} else {
				// when treating term as text query
				calledReposListNamesAll = true
				return []types.RepoName{{Name: "bar-repo"}}, nil
			}
		}
		database.Mocks.Repos.Count = mockCount
		database.Mocks.Repos.MockGetByName(t, "repo", 1)
		backend.Mocks.Repos.MockResolveRev_NoCheck(t, api.CommitID("deadbeef"))

		defer func() { database.Mocks = database.MockStores{} }()
		git.Mocks.ResolveRevision = func(rev string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
			return api.CommitID("deadbeef"), nil
		}
		defer git.ResetMocks()

		calledSearchFilesInRepos := atomic.NewBool(false)
		run.MockSearchFilesInRepos = func(args *search.TextParameters) ([]*result.FileMatch, *streaming.Stats, error) {
			calledSearchFilesInRepos.Store(true)
			if want := "foo"; args.PatternInfo.Pattern != want {
				t.Errorf("got %q, want %q", args.PatternInfo.Pattern, want)
			}
			fm := mkFileMatch(types.RepoName{Name: "repo"}, "dir/file")
			rev := "rev"
			fm.CommitID = "rev"
			fm.InputRev = &rev
			return []*result.FileMatch{fm}, &streaming.Stats{}, nil
		}
		defer func() { run.MockSearchFilesInRepos = nil }()
		for _, v := range searchVersions {
			testSuggestions(t, "foo", v, []string{"repo:foo-repo", "file:dir/file"})
			if !calledReposListNamesAll {
				t.Error("!calledReposListNamesAll")
			}
			if !calledReposListFoo {
				t.Error("!calledReposListFoo")
			}
			if !calledSearchFilesInRepos.Load() {
				t.Error("!calledSearchFilesInRepos")
			}
		}
	})

	t.Run("repo: field", func(t *testing.T) {
		var mu sync.Mutex

		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		calledReposListRepoNames := false
		database.Mocks.Repos.ListRepoNames = func(_ context.Context, op database.ReposListOptions) ([]types.RepoName, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposListRepoNames = true

			assertEqual(t, op.IncludePatterns, []string{"foo"})

			return []types.RepoName{{Name: "foo-repo"}}, nil
		}
		database.Mocks.Repos.Count = mockCount
		defer func() { database.Mocks.Repos.ListRepoNames = nil }()

		// Mock to bypass language suggestions.
		mockShowLangSuggestions = func() ([]SearchSuggestionResolver, error) { return nil, nil }
		defer func() { mockShowLangSuggestions = nil }()

		calledSearchFilesInRepos := atomic.NewBool(false)
		run.MockSearchFilesInRepos = func(args *search.TextParameters) ([]*result.FileMatch, *streaming.Stats, error) {
			mu.Lock()
			defer mu.Unlock()
			calledSearchFilesInRepos.Store(true)
			repos, err := getRepos(context.Background(), args.RepoPromise)
			if err != nil {
				t.Error(err)
			}
			if want := "foo-repo"; len(repos) != 1 || string(repos[0].Repo.Name) != want {
				t.Errorf("got %q, want %q", repos, want)
			}
			return []*result.FileMatch{
				mkFileMatch(types.RepoName{Name: "foo-repo"}, "dir/file"),
			}, &streaming.Stats{}, nil
		}
		defer func() { run.MockSearchFilesInRepos = nil }()

		for _, v := range searchVersions {
			testSuggestions(t, "repo:foo", v, []string{"repo:foo-repo", "file:dir/file"})
			if !calledReposListRepoNames {
				t.Error("!calledReposListRepoNames")
			}
			if !calledSearchFilesInRepos.Load() {
				t.Error("!calledSearchFilesInRepos")
			}
		}
	})

	t.Run("repo: field for language suggestions", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		database.Mocks.Repos.List = func(_ context.Context, have database.ReposListOptions) ([]*types.Repo, error) {
			want := database.ReposListOptions{
				IncludePatterns: []string{"foo"},
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			}
			if diff := cmp.Diff(have, want); diff != "" {
				t.Error(diff)
			}
			return []*types.Repo{{Name: "foo-repo"}}, nil
		}
		database.Mocks.Repos.ListRepoNames = func(_ context.Context, have database.ReposListOptions) ([]types.RepoName, error) {
			want := database.ReposListOptions{
				IncludePatterns: []string{"foo"},
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			}
			if diff := cmp.Diff(have, want); diff != "" {
				t.Error(diff)
			}
			return []types.RepoName{{Name: "foo-repo"}}, nil
		}
		database.Mocks.Repos.Count = mockCount
		defer func() { database.Mocks.Repos.List = nil }()
		defer func() { database.Mocks.Repos.ListRepoNames = nil }()
		git.Mocks.ResolveRevision = func(rev string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
			return api.CommitID("deadbeef"), nil
		}
		defer git.ResetMocks()

		calledReposGetInventory := false
		backend.Mocks.Repos.GetInventory = func(_ context.Context, _ *types.Repo, _ api.CommitID) (*inventory.Inventory, error) {
			calledReposGetInventory = true
			return &inventory.Inventory{
				Languages: []inventory.Lang{
					{Name: "Go"},
					{Name: "TypeScript"},
					{Name: "Java"},
				},
			}, nil
		}
		defer func() { backend.Mocks.Repos.GetInventory = nil }()

		// Mock to bypass other suggestions.
		mockShowRepoSuggestions = func() ([]SearchSuggestionResolver, error) { return nil, nil }
		defer func() { mockShowRepoSuggestions = nil }()
		mockShowFileSuggestions = func() ([]SearchSuggestionResolver, error) { return nil, nil }
		defer func() { mockShowFileSuggestions = nil }()
		mockShowSymbolMatches = func() ([]SearchSuggestionResolver, error) { return nil, nil }
		defer func() { mockShowSymbolMatches = nil }()

		for _, v := range searchVersions {
			testSuggestions(t, "repo:foo@master", v, []string{"lang:go", "lang:java", "lang:typescript"})
			if !calledReposGetInventory {
				t.Error("!calledReposGetInventory")
			}
		}
	})

	t.Run("repo: and file: field", func(t *testing.T) {
		var mu sync.Mutex

		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		calledReposListRepoNames := false
		database.Mocks.Repos.ListRepoNames = func(_ context.Context, op database.ReposListOptions) ([]types.RepoName, error) {
			mu.Lock()
			defer mu.Unlock()
			calledReposListRepoNames = true

			assertEqual(t, op.IncludePatterns, []string{"foo"})

			return []types.RepoName{{Name: "foo-repo"}}, nil
		}
		database.Mocks.Repos.Count = mockCount
		defer func() { database.Mocks.Repos.ListRepoNames = nil }()

		// Mock to bypass language suggestions.
		mockShowLangSuggestions = func() ([]SearchSuggestionResolver, error) { return nil, nil }
		defer func() { mockShowLangSuggestions = nil }()

		calledSearchFilesInRepos := atomic.NewBool(false)
		run.MockSearchFilesInRepos = func(args *search.TextParameters) ([]*result.FileMatch, *streaming.Stats, error) {
			mu.Lock()
			defer mu.Unlock()
			calledSearchFilesInRepos.Store(true)
			repos, err := getRepos(context.Background(), args.RepoPromise)
			if err != nil {
				t.Error(err)
			}
			if want := "foo-repo"; len(repos) != 1 || string(repos[0].Repo.Name) != want {
				t.Errorf("got %q, want %q", repos, want)
			}
			return []*result.FileMatch{mkFileMatch(types.RepoName{Name: "foo-repo"}, "dir/bar-file")}, &streaming.Stats{}, nil
		}
		defer func() { run.MockSearchFilesInRepos = nil }()

		for _, v := range searchVersions {
			testSuggestions(t, "repo:foo file:bar", v, []string{"file:dir/bar-file"})
			if !calledReposListRepoNames {
				t.Error("!calledReposListRepoNames")
			}
			if !calledSearchFilesInRepos.Load() {
				t.Error("!calledSearchFilesInRepos")
			}
		}
	})
}
