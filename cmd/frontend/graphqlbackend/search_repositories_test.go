package graphqlbackend

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func TestSearchRepositories(t *testing.T) {
	repositories := []*search.RepositoryRevisions{
		{Repo: &types.Repo{Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
		{Repo: &types.Repo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}},
	}
	mockSearchFilesInRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "foo.go",
				},
			}, false, nil
		case "foo/no-match":
			return []*fileMatchResolver{}, false, nil
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	// Search for all repositories
	q, err := query.ParseAndCheck("type:repo")
	if err != nil {
		t.Fatal(err)
	}
	args := search.Args{
		Pattern: &search.PatternInfo{Pattern: "", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
		Repos:   repositories,
		Query:   q,
	}
	res, _, err := searchRepositories(context.Background(), &args, int32(100))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != len(repositories) {
		t.Errorf("expected all repository results, but got %v", len(res))
	}

	// Search for all repositories where the repo name includes "one"
	q, err = query.ParseAndCheck("type:repo one")
	if err != nil {
		t.Fatal(err)
	}
	args = search.Args{
		Pattern: &search.PatternInfo{Pattern: "one", IsRegExp: true, FileMatchLimit: 1, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
		Repos:   repositories,
		Query:   q,
	}
	res, _, err = searchRepositories(context.Background(), &args, int32(100))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Errorf("expected only one repository result `foo/one`, but got %v", len(res))
	}
	if res[0].repo.repo.Name != "foo/one" {
		t.Errorf("expected the repository result to be `foo/one`, but got another repo")
	}

	// Search for all repositories where the repo name includes "foo" and the repo has a file path matching "one"
	q, err = query.ParseAndCheck("foo type:repo repohasfile:one")
	if err != nil {
		t.Fatal(err)
	}
	args = search.Args{
		Pattern: &search.PatternInfo{Pattern: "foo", IsRegExp: true, FileMatchLimit: 1, FilePatternsReposMustInclude: []string{"foo"}, PathPatternsAreRegExps: true, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true},
		Repos:   repositories,
		Query:   q,
	}
	res, _, err = searchRepositories(context.Background(), &args, int32(100))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Errorf("expected only one repository result `foo/one`, but got %v", len(res))
	}
	if res[0].repo.repo.Name != "foo/one" {
		t.Errorf("expected the repository result to be `foo/one`, but got another repo")
	}
}
