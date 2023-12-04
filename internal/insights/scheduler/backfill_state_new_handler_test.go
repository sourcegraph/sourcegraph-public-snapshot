package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Test_MovesBackfillFromNewToProcessing(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefaultReturn([]*itypes.Repo{{ID: 1, Name: "repo1"}, {ID: 2, Name: "repo2"}}, nil)
	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)
	insightsStore := store.NewInsightStore(insightsDB)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	seriesStore := store.New(insightsDB, permStore)
	repoQueryExecutor := NewMockRepoQueryExecutor()
	repoQueryExecutor.ExecuteRepoListFunc.SetDefaultReturn(nil, errors.New("repo query executor should not be called"))

	config := JobMonitorConfig{
		InsightsDB:        insightsDB,
		RepoStore:         repos,
		ObservationCtx:    &observation.TestContext,
		CostAnalyzer:      priority.NewQueryAnalyzer(),
		InsightStore:      seriesStore,
		RepoQueryExecutor: repoQueryExecutor,
	}
	var err error
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

	err = enqueueBackfill(ctx, bfs.Handle(), backfill)
	require.NoError(t, err)

	newDequeue, _, err := monitor.newBackfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	handler := newBackfillHandler{
		workerStore:     monitor.newBackfillStore,
		backfillStore:   bfs,
		seriesReader:    store.NewInsightStore(insightsDB),
		repoIterator:    discovery.NewSeriesRepoIterator(nil, repos, repoQueryExecutor),
		costAnalyzer:    *config.CostAnalyzer,
		timeseriesStore: seriesStore,
	}
	err = handler.Handle(ctx, logger, newDequeue)
	require.NoError(t, err)

	_, dupFound, err := monitor.newBackfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if dupFound {
		t.Fatal(errors.New("found record that should not be visible to the new backfill store"))
	}

	// now ensure the in progress handler _can_ pick it up
	inProgressDequeue, inProgressFound, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !inProgressFound {
		t.Fatal(errors.New("no queued record found"))
	}
	require.Equal(t, backfill.Id, inProgressDequeue.backfillId)

	recordingTimes, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, series.ID, store.SeriesPointsOpts{})
	require.NoError(t, err)
	if len(recordingTimes.RecordingTimes) == 0 {
		t.Fatal(errors.New("recording times should have been saved after success"))
	}
}

func Test_MovesBackfillFromNewToProcessing_ScopedInsight(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefaultReturn([]*itypes.Repo{}, errors.New("the repo store should not be called"))
	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)
	insightsStore := store.NewInsightStore(insightsDB)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	seriesStore := store.New(insightsDB, permStore)
	repoQueryExecutor := NewMockRepoQueryExecutor()
	repoQueryExecutor.ExecuteRepoListFunc.SetDefaultReturn([]itypes.MinimalRepo{{Name: "sourcegraph/sourcegraph", ID: 1}}, nil)

	config := JobMonitorConfig{
		InsightsDB:        insightsDB,
		RepoStore:         repos,
		ObservationCtx:    &observation.TestContext,
		CostAnalyzer:      priority.NewQueryAnalyzer(),
		InsightStore:      seriesStore,
		RepoQueryExecutor: repoQueryExecutor,
	}
	var err error
	monitor := NewBackgroundJobMonitor(ctx, config)

	repoCriteria := "repo:sourcegraph"
	series, err := insightsStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		RepositoryCriteria:  &repoCriteria,
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	require.NoError(t, err)

	backfill, err := bfs.NewBackfill(ctx, series)
	require.NoError(t, err)

	err = enqueueBackfill(ctx, bfs.Handle(), backfill)
	require.NoError(t, err)

	newDequeue, _, err := monitor.newBackfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	handler := newBackfillHandler{
		workerStore:     monitor.newBackfillStore,
		backfillStore:   bfs,
		seriesReader:    store.NewInsightStore(insightsDB),
		repoIterator:    discovery.NewSeriesRepoIterator(nil, repos, repoQueryExecutor),
		costAnalyzer:    *config.CostAnalyzer,
		timeseriesStore: seriesStore,
	}
	err = handler.Handle(ctx, logger, newDequeue)
	require.NoError(t, err)

	_, dupFound, err := monitor.newBackfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if dupFound {
		t.Fatal(errors.New("found record that should not be visible to the new backfill store"))
	}

	// now ensure the in progress handler _can_ pick it up
	inProgressDequeue, inProgressFound, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !inProgressFound {
		t.Fatal(errors.New("no queued record found"))
	}
	require.Equal(t, backfill.Id, inProgressDequeue.backfillId)

	recordingTimes, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, series.ID, store.SeriesPointsOpts{})
	require.NoError(t, err)
	if len(recordingTimes.RecordingTimes) == 0 {
		t.Fatal(errors.New("recording times should have been saved after success"))
	}
}
