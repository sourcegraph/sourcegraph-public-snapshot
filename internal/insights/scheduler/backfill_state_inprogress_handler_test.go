package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type noopBackfillRunner struct{}

func (n *noopBackfillRunner) Run(ctx context.Context, req pipeline.BackfillRequest) error {
	return nil
}

type delegateBackfillRunner struct {
	mu          sync.Mutex
	doSomething func(ctx context.Context, req pipeline.BackfillRequest) error
}

func (e *delegateBackfillRunner) Run(ctx context.Context, req pipeline.BackfillRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.doSomething(ctx, req)
}

func Test_MovesBackfillFromProcessingToComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
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
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     &noopBackfillRunner{},
		config:             newHandlerConfig(),

		clock: clock,
	}
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	_, found, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if found {
		t.Fatal(errors.New("found record that should not be visible to the new backfill store"))
	}

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
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

func Test_PullsByEstimatedCostAge(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
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

	bf1 := addBackfillToState(series, []int32{1, 2}, 3, BackfillStateProcessing)
	bf2 := addBackfillToState(series, []int32{1, 2}, 3, BackfillStateProcessing)
	bf3 := addBackfillToState(series, []int32{1, 2}, 40, BackfillStateProcessing)
	bf4 := addBackfillToState(series, []int32{1, 2}, 10, BackfillStateProcessing)

	dequeue1, _, _ := monitor.inProgressStore.Dequeue(ctx, "test1", nil)
	dequeue2, _, _ := monitor.inProgressStore.Dequeue(ctx, "test2", nil)
	dequeue3, _, _ := monitor.inProgressStore.Dequeue(ctx, "test3", nil)
	dequeue4, _, _ := monitor.inProgressStore.Dequeue(ctx, "test4", nil)

	assert.Equal(t, bf1.Id, dequeue1.backfillId)
	assert.Equal(t, bf2.Id, dequeue2.backfillId)
	assert.Equal(t, bf4.Id, dequeue3.backfillId)
	assert.Equal(t, bf3.Id, dequeue4.backfillId)
}

func Test_BackfillWithRetry(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
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
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     runner,
		config:             newHandlerConfig(),
		clock:              clock,
	}

	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateProcessing {
		t.Fatal(errors.New("backfill should be state in progress"))
	}
}

func Test_BackfillWithRetryAndComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
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
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     runner,
		config:             newHandlerConfig(),
		clock:              clock,
	}

	// we should get an errored record here that will be retried by the overall queue
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	// the queue won't immediately dequeue so we will just pass it back to the handler as if it was dequeued again
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state completed"))
	}
}

func Test_BackfillWithRepoNotFound(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(ctx context.Context, ri api.RepoID) (*itypes.Repo, error) {
		if ri == 1 {
			return &itypes.Repo{ID: 1, Name: "repo1"}, nil
		}
		return nil, &database.RepoNotFoundErr{ID: ri}
	})
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservationCtx: &observation.TestContext,
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

	runner := &delegateBackfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
			return nil
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     runner,
		config:             newHandlerConfig(),
		clock:              clock,
	}

	// we should not get an error because it's a repo not found error
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)

	it, err := completedBackfill.repoIterator(ctx, bfs)
	require.NoError(t, err)
	require.Equal(t, 0, it.ErroredRepos())

	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state completed"))
	}
}

func Test_BackfillWithARepoError(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(ctx context.Context, ri api.RepoID) (*itypes.Repo, error) {
		if ri == 1 {
			return &itypes.Repo{ID: 1, Name: "repo1"}, nil
		}
		return nil, errors.New("some error")
	})
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBackfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservationCtx: &observation.TestContext,
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

	runner := &delegateBackfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
			return nil
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     runner,
		config:             newHandlerConfig(),
		clock:              clock,
	}

	// The handler should not error
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	it, err := completedBackfill.repoIterator(ctx, bfs)
	require.NoError(t, err)

	require.Equal(t, 1, it.ErroredRepos())
	if completedBackfill.State == BackfillStateCompleted {
		t.Fatal(errors.New("backfill should not be completed"))
	}
}

func Test_BackfillWithInterrupt(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
		BackfillRunner: &noopBackfillRunner{},
		CostAnalyzer:   priority.NewQueryAnalyzer(),
	}
	monitor := NewBackgroundJobMonitor(ctx, config)

	series, err := insightsStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2", "repo3", "repo4"},
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	require.NoError(t, err)

	backfill, err := bfs.NewBackfill(ctx, series)
	require.NoError(t, err)
	backfill, err = backfill.SetScope(ctx, bfs, []int32{1, 2, 3, 4}, 0)
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
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     &runner,
		config:             newHandlerConfig(),
		clock:              clock,
	}
	handler.config.interruptAfter = time.Second * 5
	handler.config.pageSize = 2 // setting the page size to only complete 1/2 repos in 1 iteration

	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	// we will check that it was interrupted by verifying the backfill has progress, but is not completed yet
	reloaded, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	require.Equal(t, BackfillStateProcessing, reloaded.State)
	itr, err := iterator.LoadWithClock(ctx, basestore.NewWithHandle(insightsDB.Handle()), reloaded.repoIteratorId, clock)
	require.NoError(t, err)
	require.Greater(t, itr.PercentComplete, float64(0))

	// the queue won't immediately dequeue so we will just pass it back to the handler as if it was dequeued again
	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBackfill, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	if completedBackfill.State != BackfillStateCompleted {
		t.Fatal(errors.New("backfill should be state completed"))
	}
}

func Test_BackfillCrossingErrorThreshold(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
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
		ObservationCtx: &observation.TestContext,
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
	backfill, err = backfill.SetScope(ctx, bfs, []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}, 0)
	require.NoError(t, err)
	err = backfill.setState(ctx, bfs, BackfillStateProcessing)
	require.NoError(t, err)

	err = enqueueBackfill(ctx, bfs.Handle(), backfill)
	require.NoError(t, err)

	wantErr := errors.New("threshold-fake-err")

	runner := delegateBackfillRunner{doSomething: func(ctx context.Context, req pipeline.BackfillRequest) error {
		clock.Advance(time.Second * 6) // this will cause an interrupt on each iteration with a 5 second interrupt
		return wantErr
	}}

	handlerConfig := newHandlerConfig()
	handlerConfig.errorThresholdFloor = 3 // set this low enough that we will exceed it
	handlerConfig.interruptAfter = time.Hour * 24

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	handler := inProgressHandler{
		workerStore:        monitor.newBackfillStore,
		backfillStore:      bfs,
		seriesReadComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		backfillRunner:     &runner,
		config:             handlerConfig,
		clock:              clock,
	}

	err = handler.Handle(ctx, logger, dequeue)
	require.NoError(t, err)

	// we will check that it was interrupted by verifying the backfill has progress, but is not completed yet
	reloaded, err := bfs.LoadBackfill(ctx, backfill.Id)
	require.NoError(t, err)
	require.Equal(t, BackfillStateFailed, reloaded.State)
	itr, err := iterator.LoadWithClock(ctx, basestore.NewWithHandle(insightsDB.Handle()), reloaded.repoIteratorId, clock)
	require.NoError(t, err)
	require.Equal(t, itr.PercentComplete, float64(1))

	// check for incomplete points
	incomplete, err := seriesStore.LoadAggregatedIncompleteDatapoints(ctx, series.ID)
	require.NoError(t, err)
	require.Len(t, incomplete, 12)
	require.Equal(t, incomplete[0].Reason, store.ReasonExceedsErrorLimit)
}

func Test_calculateErrorThreshold(t *testing.T) {
	tests := []struct {
		name    string
		want    int
		floor   int
		percent float64
		size    int
	}{
		{
			name:    "test floor overrides percent",
			want:    10,
			floor:   10,
			percent: .05,
			size:    100,
		},
		{
			name:    "test percent overrides floor",
			want:    15,
			floor:   10,
			percent: .10,
			size:    150,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, calculateErrorThreshold(test.percent, test.floor, test.size))
		})
	}
}
