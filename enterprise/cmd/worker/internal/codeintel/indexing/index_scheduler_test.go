package indexing

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	indexSchedulerEnabled = func() bool { return true }
}

func TestIndexSchedulerUpdate(t *testing.T) {
	// TODO: Switch to store method. Right now using global hax
	var mu sync.Mutex
	searchrepos.MockResolveRepoGroups = func() (map[string][]searchrepos.RepoGroupValue, error) {
		mu.Lock()
		defer mu.Unlock()
		return map[string][]searchrepos.RepoGroupValue{
			"cncf": {
				searchrepos.RepoPath("foo-repo1"),
				searchrepos.RepoPath("repo3"),
			},
		}, nil
	}
	defer func() { searchrepos.MockResolveRepoGroups = nil }()

	database.Mocks.Repos.ListRepoNames = func(ctx context.Context, opt database.ReposListOptions) ([]types.RepoName, error) {
		return []types.RepoName{}, nil
	}
	defer func() { database.Mocks.Repos.ListRepoNames = nil }()

	// END HACKS

	mockDBStore := NewMockDBStore()
	mockDBStore.GetRepositoriesWithIndexConfigurationFunc.SetDefaultReturn([]int{43, 44, 45, 46}, nil)

	mockSettingStore := NewMockIndexingSettingStore()
	indexEnqueuer := NewMockIndexEnqueuer()

	mockRepoStore := NewMockIndexingRepoStore()
	mockRepoStore.ListRepoNamesFunc.SetDefaultReturn([]types.RepoName{{
		ID:   0,
		Name: "test repo",
	}}, nil)

	scheduler := &IndexScheduler{
		dbStore:       mockDBStore,
		settingStore:  mockSettingStore,
		repoStore:     mockRepoStore,
		operations:    newOperations(&observation.TestContext),
		indexEnqueuer: indexEnqueuer,
	}

	if err := scheduler.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(indexEnqueuer.QueueIndexesForRepositoryFunc.History()) != 6 {
		t.Errorf("unexpected number of calls to QueueIndexesForRepository. want=%d have=%d", 6, len(indexEnqueuer.QueueIndexesForRepositoryFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range indexEnqueuer.QueueIndexesForRepositoryFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg1)
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{41, 42, 43, 44, 45, 46}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository IDs (-want +got):\n%s", diff)
		}
	}
}

func TestDisabledAutoindexConfiguration(t *testing.T) {
	// ListRepoNames -> a, b, c, d
	// GetAutoindexDisabledRepositories -> c
	// Result: a, b, d
}
