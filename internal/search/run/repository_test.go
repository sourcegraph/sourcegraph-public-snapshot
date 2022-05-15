package run

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		MockReposContainingPath = func() ([]*result.FileMatch, error) {
			rev := "1a2b3c"
			return []*result.FileMatch{{
				File: result.File{
					Repo:     types.MinimalRepo{ID: 123, Name: repo.Repo.Name},
					InputRev: &rev,
					Path:     "foo.go",
				},
			}}, nil
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), job.RuntimeClients{Zoekt: zoekt}, repo, []string{"foo"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be true, but got false", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has repoHasFile filter ", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		MockReposContainingPath = func() ([]*result.FileMatch, error) {
			return nil, nil
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), job.RuntimeClients{Zoekt: zoekt}, repo, []string{"foo"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo shouldn't be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{ID: 123, Name: "foo/one"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		MockReposContainingPath = func() ([]*result.FileMatch, error) {
			rev := "1a2b3c"
			return []*result.FileMatch{{
				File: result.File{
					Repo:     types.MinimalRepo{ID: 123, Name: repo.Repo.Name},
					InputRev: &rev,
					Path:     "foo.go",
				},
			}}, nil
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), job.RuntimeClients{Zoekt: zoekt}, repo, nil, []string{"foo"})
		if err != nil {
			t.Fatal(err)
		}
		if shouldBeAdded {
			t.Errorf("Expected shouldBeAdded for repo %v to be false, but got true", repo)
		}
	})

	t.Run("repo should be included in results, query has -repoHasFile filter", func(t *testing.T) {
		repo := &search.RepositoryRevisions{Repo: types.MinimalRepo{Name: "foo/no-match"}, Revs: []search.RevisionSpecifier{{RevSpec: ""}}}
		MockReposContainingPath = func() ([]*result.FileMatch, error) {
			return nil, nil
		}
		shouldBeAdded, err := repoShouldBeAdded(context.Background(), job.RuntimeClients{Zoekt: zoekt}, repo, nil, []string{"foo"})
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
func repoShouldBeAdded(ctx context.Context, clients job.RuntimeClients, repo *search.RepositoryRevisions, filePatternsInclude, filePatternsExclude []string) (bool, error) {
	repos := []*search.RepositoryRevisions{repo}
	s := RepoSearchJob{
		FilePatternsReposMustInclude: filePatternsInclude,
		FilePatternsReposMustExclude: filePatternsExclude,
	}
	rsta, err := s.reposToAdd(ctx, clients, repos)
	if err != nil {
		return false, err
	}
	return len(rsta) == 1, nil
}
