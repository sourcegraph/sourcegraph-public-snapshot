package indexabilityscheduler

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func TestUpdate(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockDB.RepoUsageStatisticsFunc.SetDefaultReturn([]db.RepoUsageStatistics{
		{RepositoryID: 1, SearchCount: 200, PreciseCount: 50},
		{RepositoryID: 2, SearchCount: 150, PreciseCount: 25},
		{RepositoryID: 3, SearchCount: 100, PreciseCount: 35},
		{RepositoryID: 4, SearchCount: 50, PreciseCount: 100},
	}, nil)

	mockGitserverClient := gitservermocks.NewMockClient()
	mockGitserverClient.FileExistsFunc.SetDefaultHook(func(ctx context.Context, db db.DB, repositoryID int, commit, file string) (bool, error) {
		return repositoryID%2 == 0, nil
	})
	mockGitserverClient.HeadFunc.SetDefaultHook(func(ctx context.Context, db db.DB, repositoryID int) (string, error) {
		return fmt.Sprintf("c%d", repositoryID), nil
	})

	scheduler := &Scheduler{
		db:              mockDB,
		gitserverClient: mockGitserverClient,
		metrics:         NewSchedulerMetrics(metrics.TestRegisterer),
	}

	if err := scheduler.update(context.Background()); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockGitserverClient.FileExistsFunc.History()) != 4 {
		t.Errorf("unexpected number of calls to FileExists. want=%d have=%d", 2, len(mockGitserverClient.FileExistsFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range mockGitserverClient.FileExistsFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg2)
			expectedCommit := fmt.Sprintf("c%d", call.Arg2)

			if call.Arg3 != expectedCommit {
				t.Errorf("unexpected commit argument. want=%q have=%q", expectedCommit, call.Arg3)
			}
			if call.Arg4 != "go.mod" {
				t.Errorf("unexpected file argument. want=%q have=%q", "go.mod", call.Arg4)
			}
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{1, 2, 3, 4}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
		}
	}

	if len(mockDB.UpdateIndexableRepositoryFunc.History()) != 2 {
		t.Errorf("unexpected number of calls to UpdateIndexableRepository. want=%d have=%d", 2, len(mockDB.UpdateIndexableRepositoryFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range mockDB.UpdateIndexableRepositoryFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg1.RepositoryID)
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{2, 4}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
		}
	}
}
