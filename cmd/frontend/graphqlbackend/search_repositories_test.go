package graphqlbackend

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	searchbackend "github.com/sourcegraph/sourcegraph/pkg/search/backend"
)

func TestSearchRepositories(t *testing.T) {
	repositories := []*search.RepositoryRevisions{
		{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/one"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
		{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/no-match"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{}}

	mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
		repoName := args.Repos[0].Repo.Name
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?1a2b3c#" + "foo.go",
				},
			}, &searchResultsCommon{}, nil
		case "foo/no-match":
			return []*fileMatchResolver{}, &searchResultsCommon{}, nil
		default:
			return nil, &searchResultsCommon{}, errors.New("Unexpected repo")
		}
	}
	t.Run("search for all repositories", func(t *testing.T) {
		q, err := query.ParseAndCheck("type:repo")
		if err != nil {
			t.Fatal(err)
		}
		args := search.Args{
			Pattern: &search.PatternInfo{Pattern: "", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:   repositories,
			Query:   q,
			Zoekt:   zoekt,
		}
		res, _, err := searchRepositories(context.Background(), &args, int32(100))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != len(repositories) {
			t.Errorf("expected all repository results, but got %v", len(res))
		}

	})

	t.Run("search for all repositories where the repo name includes 'one'", func(t *testing.T) {
		q, err := query.ParseAndCheck("type:repo one")
		if err != nil {
			t.Fatal(err)
		}
		args := search.Args{
			Pattern: &search.PatternInfo{Pattern: "one", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:   repositories,
			Query:   q,
			Zoekt:   zoekt,
		}
		res, _, err := searchRepositories(context.Background(), &args, int32(100))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 {
			t.Errorf("expected only one repository result `foo/one`, but got %v", len(res))
		}
		r, ok := res[0].ToRepository()
		if !ok {
			t.Fatalf("expected repo result")
		}
		if r.repo.Name != "foo/one" {
			t.Errorf("expected the repository result to be `foo/one`, but got another repo")
		}
	})

	t.Run("search for all repositories where the repo name includes 'foo' and the repo has a file path matching 'one'", func(t *testing.T) {
		q, err := query.ParseAndCheck("foo type:repo repohasfile:one")
		if err != nil {
			t.Fatal(err)
		}
		args := search.Args{
			Pattern: &search.PatternInfo{Pattern: "foo", IsRegExp: true, FileMatchLimit: 1, FilePatternsReposMustInclude: []string{"foo"}, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:   repositories,
			Query:   q,
			Zoekt:   zoekt,
		}
		res, _, err := searchRepositories(context.Background(), &args, int32(100))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 {
			t.Errorf("expected only one repository result `foo/one`, but got %v", len(res))
		}
		r, ok := res[0].ToRepository()
		if !ok {
			t.Fatalf("expected repo result")
		}
		if r.repo.Name != "foo/one" {
			t.Errorf("expected the repository result to be `foo/one`, but got another repo")
		}
	})
}

func TestRepoShouldBeAdded(t *testing.T) {
	mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
		repoName := args.Repos[0].Repo.Name
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?1a2b3c#" + "foo.go",
				},
			}, &searchResultsCommon{}, nil
		case "foo/no-match":
			return []*fileMatchResolver{}, &searchResultsCommon{}, nil
		default:
			return nil, &searchResultsCommon{}, errors.New("Unexpected repo")
		}
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{}}

	t.Run("repo should be included in results, query has repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/one"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repo.Repo.Name) + "?1a2b3c#" + "foo.go",
				},
			}, &searchResultsCommon{}, nil
		}
		pat := &search.PatternInfo{Pattern: "", FilePatternsReposMustInclude: []string{"foo"}, IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if !shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be true, but got false", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has repoHasFile filter ", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/no-match"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
			return []*fileMatchResolver{}, &searchResultsCommon{}, nil
		}
		pat := &search.PatternInfo{Pattern: "", FilePatternsReposMustInclude: []string{"foo"}, IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/one"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repo.Repo.Name) + "?1a2b3c#" + "foo.go",
				},
			}, &searchResultsCommon{}, nil
		}
		pat := &search.PatternInfo{Pattern: "", FilePatternsReposMustExclude: []string{"foo"}, IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo should be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: &types.Repo{RepoIDs: types.RepoIDs{Name: "foo/no-match"}}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.Args) (matches []*fileMatchResolver, common *searchResultsCommon, err error) {
			return []*fileMatchResolver{}, &searchResultsCommon{}, nil
		}
		pat := &search.PatternInfo{Pattern: "", FilePatternsReposMustExclude: []string{"foo"}, IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if !shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be true, but got false", repo)
		}
	})
}
