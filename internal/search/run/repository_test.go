package run

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/google/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/unindexed"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/hexops/autogold"
)

func TestRepoShouldBeAdded(t *testing.T) {
	if os.Getenv("CI") != "" {
		// #25936: Some unit tests rely on external services that break
		// in CI but not locally. They should be removed or improved.
		t.Skip("TestRepoShouldBeAdded only works in local dev and is not reliable in CI")
	}
	zoekt := &searchbackend.FakeSearcher{}

	t.Run("repo should be included in results, query has repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		unindexed.MockSearchFilesInRepos = func() ([]result.Match, *streaming.Stats, error) {
			rev := "1a2b3c"
			return []result.Match{&result.FileMatch{
				File: result.File{
					Repo:     types.MinimalRepo{ID: 123, Name: repo.Repo.Name},
					InputRev: &rev,
					Path:     "foo.go",
				},
			}}, &streaming.Stats{}, nil
		}
		pat := &search.TextPatternInfo{
			Pattern:                      "",
			FilePatternsReposMustInclude: []string{"foo"},
			IsRegExp:                     true,
			FileMatchLimit:               1,
			PathPatternsAreCaseSensitive: false,
			PatternMatchesContent:        true,
			PatternMatchesPath:           true,
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if !shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be true, but got false", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has repoHasFile filter ", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		unindexed.MockSearchFilesInRepos = func() ([]result.Match, *streaming.Stats, error) {
			return nil, &streaming.Stats{}, nil
		}
		pat := &search.TextPatternInfo{
			Pattern:                      "",
			FilePatternsReposMustInclude: []string{"foo"},
			IsRegExp:                     true,
			FileMatchLimit:               1,
			PathPatternsAreCaseSensitive: false,
			PatternMatchesContent:        true,
			PatternMatchesPath:           true,
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		unindexed.MockSearchFilesInRepos = func() ([]result.Match, *streaming.Stats, error) {
			rev := "1a2b3c"
			return []result.Match{&result.FileMatch{
				File: result.File{
					Repo:     types.MinimalRepo{ID: 123, Name: repo.Repo.Name},
					InputRev: &rev,
					Path:     "foo.go",
				},
			}}, &streaming.Stats{}, nil
		}
		pat := &search.TextPatternInfo{
			Pattern:                      "",
			FilePatternsReposMustExclude: []string{"foo"},
			IsRegExp:                     true,
			FileMatchLimit:               1,
			PathPatternsAreCaseSensitive: false,
			PatternMatchesContent:        true,
			PatternMatchesPath:           true,
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), zoekt, repo, pat)
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo should be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		unindexed.MockSearchFilesInRepos = func() ([]result.Match, *streaming.Stats, error) {
			return nil, &streaming.Stats{}, nil
		}
		pat := &search.TextPatternInfo{
			Pattern:                      "",
			FilePatternsReposMustExclude: []string{"foo"},
			IsRegExp:                     true,
			FileMatchLimit:               1,
			PathPatternsAreCaseSensitive: false,
			PatternMatchesContent:        true,
			PatternMatchesPath:           true,
		}
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
func repoShouldBeAdded(ctx context.Context, zoekt zoekt.Streamer, repo *search.RepositoryRevisions, pattern *search.TextPatternInfo) (bool, error) {
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

func Test_validRepoPattern(t *testing.T) {
	test := func(input string) string {
		_, ok := validRepoPattern(input)
		return strconv.FormatBool(ok)
	}
	autogold.Want("normal pattern", "true").Equal(t, test("ok ok"))
	autogold.Want("normal pattern with space", "true").Equal(t, test("ok @thing"))
	autogold.Want("unsupported prefix", "false").Equal(t, test("@nope"))
	autogold.Want("unsupported regexp", "false").Equal(t, test("(nope).*?(@(thing))"))
}
