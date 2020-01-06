package graphqlbackend

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestSearchRepositories(t *testing.T) {
	repositories := []*search.RepositoryRevisions{
		{Repo: &types.Repo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
		{Repo: &types.Repo{ID: 456, Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
		{Repo: &types.Repo{ID: 789, Name: "bar/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{}}

	mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
		repoName := args.Repos[0].Repo.Name
		switch repoName {
		case "foo/one":
			return []*FileMatchResolver{
				{
					uri:  "git://" + string(repoName) + "?1a2b3c#" + "f.go",
					Repo: &types.Repo{ID: 123},
				},
			}, &searchResultsCommon{}, nil
		case "bar/one":
			return []*FileMatchResolver{
				{
					uri:  "git://" + string(repoName) + "?1a2b3c#" + "f.go",
					Repo: &types.Repo{ID: 789},
				},
			}, &searchResultsCommon{}, nil
		case "foo/no-match":
			return []*FileMatchResolver{}, &searchResultsCommon{}, nil
		default:
			return nil, &searchResultsCommon{}, errors.New("Unexpected repo")
		}
	}

	t.Run("search for all repositories", func(t *testing.T) {
		q, err := query.ParseAndCheck("type:repo")
		if err != nil {
			t.Fatal(err)
		}
		args := search.TextParameters{
			PatternInfo: &search.PatternInfo{Pattern: "", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:       repositories,
			Query:       q,
			Zoekt:       zoekt,
		}
		res, _, err := searchRepositories(context.Background(), &args, int32(100))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != len(repositories) {
			t.Errorf("expected all repository results, but got %v", len(res))
		}

	})

	t.Run("search for all repositories where the repo name includes 'foo/one'", func(t *testing.T) {
		q, err := query.ParseAndCheck("type:repo foo/one")
		if err != nil {
			t.Fatal(err)
		}
		args := search.TextParameters{
			PatternInfo: &search.PatternInfo{Pattern: "foo/one", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:       repositories,
			Query:       q,
			Zoekt:       zoekt,
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

	t.Run("search for all repositories where the repo name includes 'foo' and the repo has a file path matching 'f.go'", func(t *testing.T) {
		q, err := query.ParseAndCheck("foo type:repo repohasfile:f.go")
		if err != nil {
			t.Fatal(err)
		}
		args := search.TextParameters{
			PatternInfo: &search.PatternInfo{Pattern: "foo", IsRegExp: true, FileMatchLimit: 1, FilePatternsReposMustInclude: []string{"f.go"}, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
			Repos:       repositories,
			Query:       q,
			Zoekt:       zoekt,
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
	mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
		repoName := args.Repos[0].Repo.Name
		switch repoName {
		case "foo/one":
			return []*FileMatchResolver{
				{
					uri:  "git://" + string(repoName) + "?1a2b3c#" + "foo.go",
					Repo: &types.Repo{ID: 123},
				},
			}, &searchResultsCommon{}, nil
		case "foo/no-match":
			return []*FileMatchResolver{}, &searchResultsCommon{}, nil
		default:
			return nil, &searchResultsCommon{}, errors.New("Unexpected repo")
		}
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{}}

	t.Run("repo should be included in results, query has repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: &types.Repo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
			return []*FileMatchResolver{
				{
					uri:  "git://" + string(repo.Repo.Name) + "?1a2b3c#" + "foo.go",
					Repo: &types.Repo{ID: 123},
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
		repo := &search.RepositoryRevisions{Repo: &types.Repo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
			return []*FileMatchResolver{}, &searchResultsCommon{}, nil
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
		repo := &search.RepositoryRevisions{Repo: &types.Repo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
			return []*FileMatchResolver{
				{
					uri:  "git://" + string(repo.Repo.Name) + "?1a2b3c#" + "foo.go",
					Repo: &types.Repo{ID: 123},
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
		repo := &search.RepositoryRevisions{Repo: &types.Repo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		mockSearchFilesInRepos = func(args *search.TextParameters) (matches []*FileMatchResolver, common *searchResultsCommon, err error) {
			return []*FileMatchResolver{}, &searchResultsCommon{}, nil
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

// repoShouldBeAdded determines whether a repository should be included in the result set based on whether the repository fits in the subset
// of repostiories specified in the query's `repohasfile` and `-repohasfile` fields if they exist.
func repoShouldBeAdded(ctx context.Context, zoekt *searchbackend.Zoekt, repo *search.RepositoryRevisions, pattern *search.PatternInfo) (bool, error) {
	repos := []*search.RepositoryRevisions{repo}
	args := search.TextParameters{
		PatternInfo: pattern,
		Zoekt:       zoekt,
	}
	rsta, err := reposToAdd(ctx, &args, repos)
	if err != nil {
		return false, err
	}
	return len(rsta) == 1, nil
}
