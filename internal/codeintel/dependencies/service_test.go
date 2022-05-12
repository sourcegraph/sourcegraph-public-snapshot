package dependencies

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDependencies(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitService := NewMockLocalGitService()
	lockfilesService := NewMockLockfilesService()
	syncer := NewMockSyncer()
	service := testService(mockStore, gitService, lockfilesService, syncer)

	endsWithEvenDigit := func(name string) bool {
		if name == "" {
			return false
		}
		v, err := strconv.Atoi(string(name[len(name)-1]))
		if err != nil {
			return false
		}
		return v%2 == 0
	}

	mockStore.UpsertDependencyReposFunc.SetDefaultHook(func(ctx context.Context, dependencyRepos []Repo) ([]Repo, error) {
		filtered := dependencyRepos[:0]
		for _, dependencyRepo := range dependencyRepos {
			// repo is even + commit is odd, or
			// repo is odd + commit is even
			if endsWithEvenDigit(dependencyRepo.Name) != endsWithEvenDigit(dependencyRepo.Version) {
				continue
			}

			filtered = append(filtered, dependencyRepo)
		}

		return filtered, nil
	})

	gitService.GetCommitsFunc.SetDefaultHook(func(ctx context.Context, repoCommits []api.RepoCommit, _ bool) (commits []*gitdomain.Commit, _ error) {
		for _, repoCommit := range repoCommits {
			commits = append(commits, &gitdomain.Commit{ID: repoCommit.CommitID})
		}
		return commits, nil
	})

	lockfilesService.ListDependenciesFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, rev string) ([]reposource.PackageDependency, error) {
		return []reposource.PackageDependency{
			&reposource.MavenDependency{MavenModule: &reposource.MavenModule{GroupID: "g1", ArtifactID: "a1"}, Version: fmt.Sprintf("1-%s-%s", repoName, rev)},
			&reposource.MavenDependency{MavenModule: &reposource.MavenModule{GroupID: "g2", ArtifactID: "a2"}, Version: fmt.Sprintf("2-%s-%s", repoName, rev)},
			&reposource.MavenDependency{MavenModule: &reposource.MavenModule{GroupID: "g3", ArtifactID: "a3"}, Version: fmt.Sprintf("3-%s-%s", repoName, rev)},
		}, nil
	})

	repoRevs := map[api.RepoName]types.RevSpecSet{
		api.RepoName("github.com/example/foo"): {
			api.RevSpec("deadbeef1"): struct{}{},
			api.RevSpec("deadbeef2"): struct{}{},
		},
		api.RepoName("github.com/example/bar"): {
			api.RevSpec("deadbeef3"): struct{}{},
			api.RevSpec("deadbeef4"): struct{}{},
		},
	}
	dependencies, err := service.Dependencies(ctx, repoRevs)
	if err != nil {
		t.Fatalf("unexpected error querying dependencies: %s", err)
	}

	expectedDepencies := map[api.RepoName]types.RevSpecSet{
		"maven/g1/a1": {
			"v1-github.com/example/bar-deadbeef3": struct{}{},
			"v1-github.com/example/bar-deadbeef4": struct{}{},
			"v1-github.com/example/foo-deadbeef1": struct{}{},
			"v1-github.com/example/foo-deadbeef2": struct{}{},
		},
		"maven/g2/a2": {
			"v2-github.com/example/bar-deadbeef3": struct{}{},
			"v2-github.com/example/bar-deadbeef4": struct{}{},
			"v2-github.com/example/foo-deadbeef1": struct{}{},
			"v2-github.com/example/foo-deadbeef2": struct{}{},
		},
		"maven/g3/a3": {
			"v3-github.com/example/bar-deadbeef3": struct{}{},
			"v3-github.com/example/bar-deadbeef4": struct{}{},
			"v3-github.com/example/foo-deadbeef1": struct{}{},
			"v3-github.com/example/foo-deadbeef2": struct{}{},
		},
	}
	if diff := cmp.Diff(expectedDepencies, dependencies); diff != "" {
		t.Errorf("unexpected dependencies (-want +got):\n%s", diff)
	}

	syncHistory := syncer.SyncFunc.History()
	syncedRepoNames := make([]string, 0, len(syncHistory))
	for _, call := range syncHistory {
		syncedRepoNames = append(syncedRepoNames, string(call.Arg1))
	}
	sort.Strings(syncedRepoNames)

	expectedNames := []string{
		"maven/g1/a1",
		"maven/g2/a2",
		"maven/g3/a3",
	}
	if diff := cmp.Diff(expectedNames, syncedRepoNames); diff != "" {
		t.Errorf("unexpected names (-want +got):\n%s", diff)
	}
}

func testService(store Store, gitService localGitService, lockfilesService LockfilesService, syncer Syncer) *Service {
	return newService(
		store,
		gitService,
		lockfilesService,
		semaphore.NewWeighted(100),
		syncer,
		semaphore.NewWeighted(100),
		&observation.TestContext,
	)
}
