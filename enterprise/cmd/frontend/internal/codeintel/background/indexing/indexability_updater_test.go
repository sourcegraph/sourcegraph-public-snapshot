package indexing

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/time/rate"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	indexabilityUpdaterEnabled = func() bool { return true }
}

func TestIndexabilityUpdater(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.RepoUsageStatisticsFunc.SetDefaultReturn([]store.RepoUsageStatistics{
		{RepositoryID: 1, SearchCount: 200, PreciseCount: 50},
		{RepositoryID: 2, SearchCount: 150, PreciseCount: 25},
		{RepositoryID: 3, SearchCount: 100, PreciseCount: 35},
		{RepositoryID: 4, SearchCount: 50, PreciseCount: 100},
	}, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.HeadFunc.SetDefaultHook(func(ctx context.Context, repositoryID int) (string, error) {
		return fmt.Sprintf("c%d", repositoryID), nil
	})
	mockGitserverClient.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error) {
		if repositoryID%2 == 0 {
			return []string{"go.mod"}, nil
		}
		return nil, nil
	})

	updater := &IndexabilityUpdater{
		dbStore:            mockDBStore,
		gitserverClient:    mockGitserverClient,
		operations:         newOperations(&observation.TestContext),
		limiter:            rate.NewLimiter(MaxGitserverRequestsPerSecond, 1),
		enableIndexingCNCF: false,
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockGitserverClient.ListFilesFunc.History()) != 4 {
		t.Errorf("unexpected number of calls to ListFiles. want=%d have=%d", 2, len(mockGitserverClient.ListFilesFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range mockGitserverClient.ListFilesFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg1)
			expectedCommit := fmt.Sprintf("c%d", call.Arg1)

			if call.Arg2 != expectedCommit {
				t.Errorf("unexpected commit argument. want=%q have=%q", expectedCommit, call.Arg2)
			}
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{1, 2, 3, 4}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
		}
	}

	if len(mockDBStore.UpdateIndexableRepositoryFunc.History()) != 2 {
		t.Errorf("unexpected number of calls to UpdateIndexableRepository. want=%d have=%d", 2, len(mockDBStore.UpdateIndexableRepositoryFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range mockDBStore.UpdateIndexableRepositoryFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg1.RepositoryID)
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{2, 4}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
		}
	}

	if len(mockDBStore.ResetIndexableRepositoriesFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to ResetIndexableRepositories. want=%d have=%d", 2, len(mockDBStore.ResetIndexableRepositoriesFunc.History()))
	}
}

func TestSkipManualUploads(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.RepoUsageStatisticsFunc.SetDefaultReturn([]store.RepoUsageStatistics{
		{RepositoryID: 1, SearchCount: 200, PreciseCount: 50},
	}, nil)
	mockDBStore.GetUploadsFunc.SetDefaultReturn([]store.Upload{
		{AssociatedIndexID: nil},
	}, 1, nil)
	mockGitserverClient := NewMockGitserverClient()

	updater := &IndexabilityUpdater{
		dbStore:            mockDBStore,
		gitserverClient:    mockGitserverClient,
		operations:         newOperations(&observation.TestContext),
		limiter:            rate.NewLimiter(MaxGitserverRequestsPerSecond, 1),
		enableIndexingCNCF: false,
	}

	err := updater.Handle(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(mockDBStore.UpdateIndexableRepositoryFunc.history) > 0 {
		t.Fatal("IndexabilityUpdater tried to queue index for repo with recent manual upload")
	}
}
