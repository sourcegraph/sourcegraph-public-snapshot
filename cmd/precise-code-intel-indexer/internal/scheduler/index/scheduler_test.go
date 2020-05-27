package indexscheduler

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
	mockDB.TransactFunc.SetDefaultReturn(mockDB, nil)
	mockDB.IndexableRepositoriesFunc.SetDefaultReturn([]db.IndexableRepository{
		{RepositoryID: 1},
		{RepositoryID: 2},
		{RepositoryID: 3},
		{RepositoryID: 4},
	}, nil)
	mockDB.IsQueuedFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) (bool, error) {
		return repositoryID%2 != 0, nil
	})

	mockGitserverClient := gitservermocks.NewMockClient()
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

	if len(mockDB.IsQueuedFunc.History()) != 4 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 4, len(mockDB.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDB.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c1", "c2", "c3", "c4"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}

	if len(mockDB.InsertIndexFunc.History()) != 2 {
		t.Errorf("unexpected number of calls to InsertIndex. want=%d have=%d", 2, len(mockDB.InsertIndexFunc.History()))
	} else {
		indexCommits := map[int]string{}
		for _, call := range mockDB.InsertIndexFunc.History() {
			indexCommits[call.Arg1.RepositoryID] = call.Arg1.Commit
		}

		expectedIndexCommits := map[int]string{
			2: "c2",
			4: "c4",
		}
		if diff := cmp.Diff(expectedIndexCommits, indexCommits); diff != "" {
			t.Errorf("unexpected indexes (-want +got):\n%s", diff)
		}
	}

	if len(mockDB.UpdateIndexableRepositoryFunc.History()) != 2 {
		t.Errorf("unexpected number of calls to UpdateIndexableRepository. want=%d have=%d", 2, len(mockDB.UpdateIndexableRepositoryFunc.History()))
	}
}
