package dependencies

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestDependencies(t *testing.T) {
	// Ensure the precise flag is enabled
	enablePreciseQueries = true

	ctx := context.Background()
	mockStore := NewMockStore()
	gitService := NewMockLocalGitService()
	lockfilesService := NewMockLockfilesService()
	syncer := NewMockSyncer()
	service := testService(mockStore, gitService, lockfilesService, syncer)

	mockStore.PreciseDependenciesFunc.SetDefaultHook(func(ctx context.Context, repoName, commit string) (map[api.RepoName]types.RevSpecSet, error) {
		switch repoName {
		case "github.com/example/baz":
			return map[api.RepoName]types.RevSpecSet{
				api.RepoName(fmt.Sprintf("%s-depA", repoName)): {"deadbeef1": struct{}{}},
				api.RepoName(fmt.Sprintf("%s-depB", repoName)): {"deadbeef2": struct{}{}},
				api.RepoName(fmt.Sprintf("%s-depC", repoName)): {"deadbeef3": struct{}{}},
			}, nil
		case "github.com/example/quux":
			return map[api.RepoName]types.RevSpecSet{
				api.RepoName(fmt.Sprintf("%s-depA", repoName)): {"deadbeef1": struct{}{}},
				api.RepoName(fmt.Sprintf("%s-depB", repoName)): {"deadbeef2": struct{}{}},
			}, nil
		}

		return nil, nil
	})

	// Return canned dependencies for repo `baz`
	mockStore.LockfileDependenciesFunc.SetDefaultHook(func(ctx context.Context, opts store.LockfileDependenciesOpts) ([]shared.PackageDependency, bool, error) {
		if opts.RepoName != "github.com/example/baz" {
			return nil, false, nil
		}

		return []shared.PackageDependency{
			shared.TestPackageDependencyLiteral("npm/leftpad", "1", "2", "3", "4"),
			shared.TestPackageDependencyLiteral("npm/rightpad", "2", "3", "4", "5"),
			shared.TestPackageDependencyLiteral("npm/centerpad", "3", "4", "5", "6"),
		}, true, nil
	})

	// GetCommits returns the same values as input; no errors
	gitService.GetCommitsFunc.SetDefaultHook(func(ctx context.Context, repoCommits []api.RepoCommit, _ bool) (commits []*gitdomain.Commit, _ error) {
		for _, repoCommit := range repoCommits {
			commits = append(commits, &gitdomain.Commit{ID: repoCommit.CommitID})
		}
		return commits, nil
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
		api.RepoName("github.com/example/baz"): {
			api.RevSpec("deadbeef5"): struct{}{},
			api.RevSpec("deadbeef6"): struct{}{},
		},
	}
	dependencies, notFound, err := service.Dependencies(ctx, repoRevs, false)
	if err != nil {
		t.Fatalf("unexpected error querying dependencies: %s", err)
	}

	expectedDependencies := map[api.RepoName]types.RevSpecSet{
		// From precise dependencies
		"github.com/example/baz-depA": {"deadbeef1": struct{}{}},
		"github.com/example/baz-depB": {"deadbeef2": struct{}{}},
		"github.com/example/baz-depC": {"deadbeef3": struct{}{}},

		// From dependencies store
		("npm/leftpad"):   {"1": struct{}{}},
		("npm/rightpad"):  {"2": struct{}{}},
		("npm/centerpad"): {"3": struct{}{}},
	}
	if diff := cmp.Diff(expectedDependencies, dependencies); diff != "" {
		t.Errorf("unexpected dependencies (-want +got):\n%s", diff)
	}

	expectedNotFound := map[api.RepoName]types.RevSpecSet{
		api.RepoName("github.com/example/foo"): {
			api.RevSpec("deadbeef1"): struct{}{},
			api.RevSpec("deadbeef2"): struct{}{},
		},
		api.RepoName("github.com/example/bar"): {
			api.RevSpec("deadbeef3"): struct{}{},
			api.RevSpec("deadbeef4"): struct{}{},
		},
	}

	if diff := cmp.Diff(expectedNotFound, notFound); diff != "" {
		t.Errorf("unexpected notFound (-want +got):\n%s", diff)
	}

	// Located in the end so as not to interfere with Upsert call counting.
	t.Run("get-commits-error", func(t *testing.T) {
		getCommitsErr := errors.New("get commits failed for at least one commit")

		gitService.GetCommitsFunc.PushHook(func(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) (commits []*gitdomain.Commit, _ error) {
			for _, repoCommit := range repoCommits {
				if repoCommit.Repo != "github.com/example/quux" {
					if ignoreErrors {
						commits = append(commits, nil)
						continue
					}
					return nil, getCommitsErr
				}
				commits = append(commits, &gitdomain.Commit{ID: repoCommit.CommitID})
			}
			return commits, nil
		})

		repoRevs := map[api.RepoName]types.RevSpecSet{
			api.RepoName("github.com/example/foo"): {
				api.RevSpec("deadbeef1"): struct{}{},
			},
			api.RepoName("github.com/example/quux"): {
				api.RevSpec("deadbeef1"): struct{}{},
			},
		}

		dependencies, notFound, err := service.Dependencies(ctx, repoRevs, false)
		if err != nil {
			t.Fatalf("unexpected error querying dependencies: %s", err)
		}
		expectedDepencies := map[api.RepoName]types.RevSpecSet{
			"github.com/example/quux-depA": {"deadbeef1": struct{}{}},
			"github.com/example/quux-depB": {"deadbeef2": struct{}{}},
		}
		if diff := cmp.Diff(expectedDepencies, dependencies); diff != "" {
			t.Errorf("unexpected dependencies (-want +got):\n%s", diff)
		}

		expectedNotFound := map[api.RepoName]types.RevSpecSet{}
		if diff := cmp.Diff(expectedNotFound, notFound); diff != "" {
			t.Errorf("unexpected notFound (-want +got):\n%s", diff)
		}
	})
}

func TestDependents(t *testing.T) {
	// Ensure the precise flag is enabled
	enablePreciseQueries = true

	ctx := context.Background()
	mockStore := NewMockStore()
	gitService := NewMockLocalGitService()
	lockfilesService := NewMockLockfilesService()
	syncer := NewMockSyncer()
	service := testService(mockStore, gitService, lockfilesService, syncer)

	// GetCommits returns the same values as input; no errors
	gitService.GetCommitsFunc.SetDefaultHook(func(ctx context.Context, repoCommits []api.RepoCommit, _ bool) (commits []*gitdomain.Commit, _ error) {
		for _, repoCommit := range repoCommits {
			commits = append(commits, &gitdomain.Commit{ID: repoCommit.CommitID})
		}
		return commits, nil
	})

	mockStore.PreciseDependentsFunc.SetDefaultHook(func(ctx context.Context, repoName, commit string) (map[api.RepoName]types.RevSpecSet, error) {
		if repoName != "github.com/example/baz" {
			return nil, nil
		}

		return map[api.RepoName]types.RevSpecSet{
			api.RepoName(fmt.Sprintf("%s-depA", repoName)): {"deadbeef1": struct{}{}},
			api.RepoName(fmt.Sprintf("%s-depB", repoName)): {"deadbeef2": struct{}{}},
			api.RepoName(fmt.Sprintf("%s-depC", repoName)): {"deadbeef3": struct{}{}},
		}, nil
	})

	mockStore.LockfileDependentsFunc.SetDefaultHook(func(ctx context.Context, repoName, commit string) ([]api.RepoCommit, error) {
		return []api.RepoCommit{
			{Repo: api.RepoName(fmt.Sprintf("dep-a-%s", repoName)), CommitID: api.CommitID(fmt.Sprintf("c-%s", commit))},
			{Repo: api.RepoName(fmt.Sprintf("dep-b-%s", repoName)), CommitID: api.CommitID(fmt.Sprintf("c-%s", commit))},
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
		api.RepoName("github.com/example/baz"): {
			api.RevSpec("deadbeef5"): struct{}{},
			api.RevSpec("deadbeef6"): struct{}{},
		},
	}
	dependents, err := service.Dependents(ctx, repoRevs)
	if err != nil {
		t.Fatalf("unexpected error querying dependents: %s", err)
	}

	expectedDependents := map[api.RepoName]types.RevSpecSet{
		// From precise dependents
		"github.com/example/baz-depA": {"deadbeef1": struct{}{}},
		"github.com/example/baz-depB": {"deadbeef2": struct{}{}},
		"github.com/example/baz-depC": {"deadbeef3": struct{}{}},

		// From lockfile dependents
		api.RepoName("dep-a-github.com/example/foo"): {
			api.RevSpec("c-deadbeef1"): struct{}{},
			api.RevSpec("c-deadbeef2"): struct{}{},
		},
		api.RepoName("dep-a-github.com/example/bar"): {
			api.RevSpec("c-deadbeef3"): struct{}{},
			api.RevSpec("c-deadbeef4"): struct{}{},
		},
		api.RepoName("dep-a-github.com/example/baz"): {
			api.RevSpec("c-deadbeef5"): struct{}{},
			api.RevSpec("c-deadbeef6"): struct{}{},
		},
		api.RepoName("dep-b-github.com/example/foo"): {
			api.RevSpec("c-deadbeef1"): struct{}{},
			api.RevSpec("c-deadbeef2"): struct{}{},
		},
		api.RepoName("dep-b-github.com/example/bar"): {
			api.RevSpec("c-deadbeef3"): struct{}{},
			api.RevSpec("c-deadbeef4"): struct{}{},
		},
		api.RepoName("dep-b-github.com/example/baz"): {
			api.RevSpec("c-deadbeef5"): struct{}{},
			api.RevSpec("c-deadbeef6"): struct{}{},
		},
	}
	if diff := cmp.Diff(expectedDependents, dependents); diff != "" {
		t.Errorf("unexpected dependents (-want +got):\n%s", diff)
	}
}

func TestIndexLockfiles(t *testing.T) {
	// Ensure lockfile indexing is enabled
	oldLockfileIndexingEnabled := lockfileIndexingEnabled
	lockfileIndexingEnabled = func() bool { return true }
	defer func() { lockfileIndexingEnabled = oldLockfileIndexingEnabled }()

	ctx := context.Background()
	mockStore := NewMockStore()
	gitService := NewMockLocalGitService()
	lockfilesService := NewMockLockfilesService()
	syncer := NewMockSyncer()
	service := testService(mockStore, gitService, lockfilesService, syncer)

	// GetCommits returns the same values as input; no errors
	gitService.GetCommitsFunc.SetDefaultHook(func(ctx context.Context, repoCommits []api.RepoCommit, _ bool) (commits []*gitdomain.Commit, _ error) {
		for _, repoCommit := range repoCommits {
			commits = append(commits, &gitdomain.Commit{ID: repoCommit.CommitID})
		}
		return commits, nil
	})

	// Return archive dependencies for repos `foo` and `bar`
	lockfilesService.ListDependenciesFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, rev string) ([]lockfiles.Result, error) {
		if repoName != "github.com/example/foo" && repoName != "github.com/example/bar" {
			return []lockfiles.Result{}, nil
		}

		mavenPackages := []reposource.VersionedPackage{
			&reposource.MavenVersionedPackage{MavenModule: &reposource.MavenModule{GroupID: "g1", ArtifactID: "a1"}, Version: fmt.Sprintf("1-%s-%s", repoName, rev)},
			&reposource.MavenVersionedPackage{MavenModule: &reposource.MavenModule{GroupID: "g2", ArtifactID: "a2"}, Version: fmt.Sprintf("2-%s-%s", repoName, rev)},
			&reposource.MavenVersionedPackage{MavenModule: &reposource.MavenModule{GroupID: "g3", ArtifactID: "a3"}, Version: fmt.Sprintf("3-%s-%s", repoName, rev)},
		}

		graph1 := &lockfiles.DependencyGraph{}
		graph2 := &lockfiles.DependencyGraph{}

		switch rev {
		case "deadbeef1":
			return []lockfiles.Result{{Lockfile: "pom.xml", Deps: mavenPackages, Graph: graph1}}, nil
		case "deadbeef2":
			return []lockfiles.Result{
				{Lockfile: "pom.xml", Deps: mavenPackages, Graph: graph2},
				{Lockfile: "pom2.xml", Deps: mavenPackages, Graph: nil},
			}, nil
		default:
			return []lockfiles.Result{{Lockfile: "pom.xml", Deps: mavenPackages, Graph: nil}}, nil
		}
	})

	// UpsertDependencyRepos influences the value that syncer.Sync is called with (asserted below)
	mockStore.UpsertDependencyReposFunc.SetDefaultHook(func(ctx context.Context, dependencyRepos []Repo) ([]Repo, error) {
		filtered := dependencyRepos[:0]
		for _, dependencyRepo := range dependencyRepos {
			// repo is even + commit is odd, or
			// repo is odd + commit is even
			if endsWithEvenDigit(string(dependencyRepo.Name)) != endsWithEvenDigit(dependencyRepo.Version) {
				continue
			}

			filtered = append(filtered, dependencyRepo)
		}

		return filtered, nil
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
		api.RepoName("github.com/example/baz"): {
			api.RevSpec("deadbeef5"): struct{}{},
			api.RevSpec("deadbeef6"): struct{}{},
		},
	}

	err := service.IndexLockfiles(ctx, repoRevs)
	if err != nil {
		t.Fatalf("unexpected error querying dependencies: %s", err)
	}

	// Assert `store.UpsertLockfileGraph` was called
	mockassert.CalledN(t, mockStore.UpsertLockfileGraphFunc, 7)
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/foo", "deadbeef1", "pom.xml", mockassert.Skip))
	// deadbeef2 has 2 results
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/foo", "deadbeef2", "pom.xml", mockassert.Skip))
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/foo", "deadbeef2", "pom2.xml", mockassert.Skip))
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/bar", "deadbeef3", "pom.xml", mockassert.Skip))
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/bar", "deadbeef4", "pom.xml", mockassert.Skip))
	// We make sure that "0 dependencies" is also recorded as a result
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/baz", "deadbeef5", "NOT-FOUND", []shared.PackageDependency{}, nil))
	mockassert.CalledOnceWith(t, mockStore.UpsertLockfileGraphFunc, mockassert.Values(mockassert.Skip, "github.com/example/baz", "deadbeef6", "NOT-FOUND", []shared.PackageDependency{}, nil))

	// Assert `syncer.Sync` was called correctly
	syncHistory := syncer.SyncFunc.History()
	syncedRepoNames := make([]string, 0, len(syncHistory))
	for _, call := range syncHistory {
		syncedRepoNames = append(syncedRepoNames, string(call.Arg1))
	}
	sort.Strings(syncedRepoNames)

	// It should have synced the dependencies not filtered out by
	// UpsertDependencyRepos above.
	expectedNames := []string{
		"maven/g1/a1",
		"maven/g2/a2",
		"maven/g3/a3",
	}
	if diff := cmp.Diff(expectedNames, syncedRepoNames); diff != "" {
		t.Errorf("unexpected names (-want +got):\n%s", diff)
	}
}

func endsWithEvenDigit(name string) bool {
	if name == "" {
		return false
	}
	v, err := strconv.Atoi(string(name[len(name)-1]))
	if err != nil {
		return false
	}
	return v%2 == 0
}

func testService(store store.Store, gitService localGitService, lockfilesService LockfilesService, syncer Syncer) *Service {
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
