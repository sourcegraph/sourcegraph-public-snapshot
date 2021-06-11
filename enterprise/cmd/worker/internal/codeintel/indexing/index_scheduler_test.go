package indexing

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	indexSchedulerEnabled = func() bool { return true }
}

func TestIndexSchedulerUpdate(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.GetRepositoriesWithIndexConfigurationFunc.SetDefaultReturn([]int{43, 44, 45, 46}, nil)
	mockDBStore.IndexableRepositoriesFunc.SetDefaultReturn([]store.IndexableRepository{
		{RepositoryID: 41},
		{RepositoryID: 42},
		{RepositoryID: 43},
		{RepositoryID: 44},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	scheduler := &IndexScheduler{
		dbStore:       mockDBStore,
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
