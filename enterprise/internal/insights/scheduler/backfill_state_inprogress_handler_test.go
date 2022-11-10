package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

type delegateBackfillRunner struct {
	doSomething func(ctx context.Context, req pipeline.BackfillRequest) error
}

func (e *delegateBackfillRunner) Run(ctx context.Context, req pipeline.BackfillRequest) error {
	return e.doSomething(ctx, req)
}

func Test_MovesBackfillFromProcessingToComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	permStore := store.NewInsightPermissionStore(database.NewMockDB())
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
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
		seriesReader:   insightsStore,
		repoStore:      repos,
		insightsStore:  seriesStore,
		backfillRunner: &noopBackfillRunner{},
		config:         newHandlerConfig(),

		clock: clock,
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

func Test_PullsByPriorityGroupAge(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	permStore := store.NewInsightPermissionStore(database.NewMockDB())
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
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

	addBackfillToState := func(series types.InsightSeries, scope []int32, cost float64, state BackfillState) *SeriesBackfill {
		backfill, err := bfs.NewBackfill(ctx, series)
		require.NoError(t, err)
		backfill, err = backfill.SetScope(ctx, bfs, scope, cost)
		require.NoError(t, err)
		err = backfill.setState(ctx, bfs, state)
		require.NoError(t, err)

		err = enqueueBackfill(ctx, bfs.Handle(), backfill)
		require.NoError(t, err)
		return backfill
	}

	bf1 := addBackfillToState(series, []int32{1, 2}, 5, BackfillStateProcessing)
	bf2 := addBackfillToState(series, []int32{1, 2}, 3, BackfillStateProcessing)
	bf3 := addBackfillToState(series, []int32{1, 2}, 40, BackfillStateProcessing)
	bf4 := addBackfillToState(series, []int32{1, 2}, 10, BackfillStateProcessing)

	dequeue1, _, _ := monitor.inProgressStore.Dequeue(ctx, "test1", nil)
	job1 := dequeue1.(*BaseJob)
	dequeue2, _, _ := monitor.inProgressStore.Dequeue(ctx, "test2", nil)
	job2 := dequeue2.(*BaseJob)
	dequeue3, _, _ := monitor.inProgressStore.Dequeue(ctx, "test3", nil)
	job3 := dequeue3.(*BaseJob)
	dequeue4, _, _ := monitor.inProgressStore.Dequeue(ctx, "test4", nil)
	job4 := dequeue4.(*BaseJob)

	// cost split is in 4 equal buckets based on 0 - max(cost)

	// 1st job is bf1 it has higher cost but it's grouped in same cost and is older
	assert.Equal(t, bf1.Id, job1.backfillId)
	assert.Equal(t, bf2.Id, job2.backfillId)
	assert.Equal(t, bf4.Id, job3.backfillId)
	assert.Equal(t, bf3.Id, job4.backfillId)

}

func Test_BackfillWithRetry(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	permStore := store.NewInsightPermissionStore(database.NewMockDB())
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
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

	attemptCounts := make(map[int]int)
	runner := &delegateBackfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
			val := attemptCounts[int(req.Repo.ID)]
			attemptCounts[int(req.Repo.ID)] += 1
			if val > 2 {
				return nil
			}
			return errors.New("fake error")
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:    monitor.newBackfillStore,
		backfillStore:  bfs,
		seriesReader:   insightsStore,
		repoStore:      repos,
		insightsStore:  seriesStore,
		backfillRunner: runner,
		config:         newHandlerConfig(),
		clock:          clock,
	}

	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.loadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateProcessing {
		t.Fatal(errors.New("backfill should be state in progress"))
	}
}

func Test_BackfillWithRetryAndComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	permStore := store.NewInsightPermissionStore(database.NewMockDB())
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
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

	attemptCounts := make(map[int]int)
	runner := &delegateBackfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
			val := attemptCounts[int(req.Repo.ID)]
			attemptCounts[int(req.Repo.ID)] += 1
			if val > 2 {
				return nil
			}
			return errors.New("fake error")
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:    monitor.newBackfillStore,
		backfillStore:  bfs,
		seriesReader:   insightsStore,
		repoStore:      repos,
		insightsStore:  seriesStore,
		backfillRunner: runner,
		config:         newHandlerConfig(),
		clock:          clock,
	}

	// we should get an errored record here that will be retried by the overall queue
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	// the queue won't immediately dequeue so we will just pass it back to the handler as if it was dequeued again
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.loadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state completed"))
	}
}

func Test_BackfillWithInterrupt(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	permStore := store.NewInsightPermissionStore(database.NewMockDB())
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&itypes.Repo{ID: 1, Name: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObsContext:     &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
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

	runner := delegateBackfillRunner{doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
		clock.Advance(time.Second * 6) // this will cause an interrupt on each iteration with a 5 second interrupt
		return nil
	}}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:    monitor.newBackfillStore,
		backfillStore:  bfs,
		seriesReader:   insightsStore,
		repoStore:      repos,
		insightsStore:  seriesStore,
		backfillRunner: &runner,
		config:         newHandlerConfig(),
		clock:          clock,
	}
	handler.config.interruptAfter = time.Second * 5

	// we should get an errored record here that will be retried by the overall queue
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	// we will check that it was interrupted by verifying the backfill has progress, but is not completed yet
	reloaded, err := bfs.loadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	require.Equal(t, BackfillStateProcessing, reloaded.State)
	itr, err := iterator.LoadWithClock(ctx, basestore.NewWithHandle(insightsDB.Handle()), reloaded.repoIteratorId, clock)
	require.NoError(t, err)
	require.Greater(t, itr.PercentComplete, float64(0))

	// the queue won't immediately dequeue so we will just pass it back to the handler as if it was dequeued again
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.loadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state completed"))
	}
}
