package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type noopBackfillRunner struct {
}

func (n *noopBackfillRunner) Run(ctx context.Context, req pipeline.BackfillRequest) error {
	return nil
}

func Test_MovesBackfillFromProcessingToComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)
	insightsStore := store.NewInsightStore(insightsDB)
	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
	}
	monitor := NewBackgroundJobMonitor(ctx, config)

	series, err := insightsStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	require.NoError(t, err)

	backfill, err := bfs.NewBackfill(ctx, series)
	require.NoError(t, err)
	backfill, err = backfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = backfill.setState(ctx, bfs, BackfillStateProcessing)
	require.NoError(t, err)

	err = enqueueBackfill(ctx, bfs.Handle(), backfill)
	require.NoError(t, err)

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:    monitor.newBackfillStore,
		backfillStore:  bfs,
		seriesReader:   store.NewInsightStore(insightsDB),
		repoStore:      repos,
		backfillRunner: &noopBackfillRunner{},
	}
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	_, found, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if found {
		t.Fatal(errors.New("found record that should not be visible to the new backfill store"))
	}

	completedBackfill, err := bfs.loadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state COMPLETED after success"))
	}
	completedItr, err := iterator.Load(ctx, bfs.Store, completedBackfill.repoIteratorId)
	require.NoError(t, err)
	if completedItr.CompletedAt.IsZero() {
		t.Fatal(errors.New("iterator should be COMPLETED after success"))
	}

}
