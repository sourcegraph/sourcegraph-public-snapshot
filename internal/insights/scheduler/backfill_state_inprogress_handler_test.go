pbckbge scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/pipeline"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler/iterbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type noopBbckfillRunner struct{}

func (n *noopBbckfillRunner) Run(ctx context.Context, req pipeline.BbckfillRequest) error {
	return nil
}

type delegbteBbckfillRunner struct {
	mu          sync.Mutex
	doSomething func(ctx context.Context, req pipeline.BbckfillRequest) error
}

func (e *delegbteBbckfillRunner) Run(ctx context.Context, req pipeline.BbckfillRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.doSomething(ctx, req)
}

func Test_MovesBbckfillFromProcessingToComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     &noopBbckfillRunner{},
		config:             newHbndlerConfig(),

		clock: clock,
	}
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	_, found, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if found {
		t.Fbtbl(errors.New("found record thbt should not be visible to the new bbckfill store"))
	}

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	if completedBbckfill.Stbte != BbckfillStbteCompleted {
		t.Fbtbl(errors.New("bbckfill should be stbte COMPLETED bfter success"))
	}
	completedItr, err := iterbtor.Lobd(ctx, bfs.Store, completedBbckfill.repoIterbtorId)
	require.NoError(t, err)
	if completedItr.CompletedAt.IsZero() {
		t.Fbtbl(errors.New("iterbtor should be COMPLETED bfter success"))
	}
}

func Test_PullsByEstimbtedCostAge(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bddBbckfillToStbte := func(series types.InsightSeries, scope []int32, cost flobt64, stbte BbckfillStbte) *SeriesBbckfill {
		bbckfill, err := bfs.NewBbckfill(ctx, series)
		require.NoError(t, err)
		bbckfill, err = bbckfill.SetScope(ctx, bfs, scope, cost)
		require.NoError(t, err)
		err = bbckfill.setStbte(ctx, bfs, stbte)
		require.NoError(t, err)

		err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
		require.NoError(t, err)
		return bbckfill
	}

	bf1 := bddBbckfillToStbte(series, []int32{1, 2}, 3, BbckfillStbteProcessing)
	bf2 := bddBbckfillToStbte(series, []int32{1, 2}, 3, BbckfillStbteProcessing)
	bf3 := bddBbckfillToStbte(series, []int32{1, 2}, 40, BbckfillStbteProcessing)
	bf4 := bddBbckfillToStbte(series, []int32{1, 2}, 10, BbckfillStbteProcessing)

	dequeue1, _, _ := monitor.inProgressStore.Dequeue(ctx, "test1", nil)
	dequeue2, _, _ := monitor.inProgressStore.Dequeue(ctx, "test2", nil)
	dequeue3, _, _ := monitor.inProgressStore.Dequeue(ctx, "test3", nil)
	dequeue4, _, _ := monitor.inProgressStore.Dequeue(ctx, "test4", nil)

	bssert.Equbl(t, bf1.Id, dequeue1.bbckfillId)
	bssert.Equbl(t, bf2.Id, dequeue2.bbckfillId)
	bssert.Equbl(t, bf4.Id, dequeue3.bbckfillId)
	bssert.Equbl(t, bf3.Id, dequeue4.bbckfillId)
}

func Test_BbckfillWithRetry(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	bttemptCounts := mbke(mbp[int]int)
	runner := &delegbteBbckfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {

			vbl := bttemptCounts[int(req.Repo.ID)]
			bttemptCounts[int(req.Repo.ID)] += 1
			if vbl > 2 {
				return nil
			}
			return errors.New("fbke error")
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     runner,
		config:             newHbndlerConfig(),
		clock:              clock,
	}

	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	if completedBbckfill.Stbte != BbckfillStbteProcessing {
		t.Fbtbl(errors.New("bbckfill should be stbte in progress"))
	}
}

func Test_BbckfillWithRetryAndComplete(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	bttemptCounts := mbke(mbp[int]int)
	runner := &delegbteBbckfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {
			vbl := bttemptCounts[int(req.Repo.ID)]
			bttemptCounts[int(req.Repo.ID)] += 1
			if vbl > 2 {
				return nil
			}
			return errors.New("fbke error")
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     runner,
		config:             newHbndlerConfig(),
		clock:              clock,
	}

	// we should get bn errored record here thbt will be retried by the overbll queue
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	// the queue won't immedibtely dequeue so we will just pbss it bbck to the hbndler bs if it wbs dequeued bgbin
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	if completedBbckfill.Stbte != BbckfillStbteCompleted {
		t.Fbtbl(errors.New("bbckfill should be stbte completed"))
	}
}

func Test_BbckfillWithRepoNotFound(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(ctx context.Context, ri bpi.RepoID) (*itypes.Repo, error) {
		if ri == 1 {
			return &itypes.Repo{ID: 1, Nbme: "repo1"}, nil
		}
		return nil, &dbtbbbse.RepoNotFoundErr{ID: ri}
	})
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	runner := &delegbteBbckfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {
			return nil
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     runner,
		config:             newHbndlerConfig(),
		clock:              clock,
	}

	// we should not get bn error becbuse it's b repo not found error
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)

	it, err := completedBbckfill.repoIterbtor(ctx, bfs)
	require.NoError(t, err)
	require.Equbl(t, 0, it.ErroredRepos())

	if completedBbckfill.Stbte != BbckfillStbteCompleted {
		t.Fbtbl(errors.New("bbckfill should be stbte completed"))
	}
}

func Test_BbckfillWithARepoError(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(ctx context.Context, ri bpi.RepoID) (*itypes.Repo, error) {
		if ri == 1 {
			return &itypes.Repo{ID: 1, Nbme: "repo1"}, nil
		}
		return nil, errors.New("some error")
	})
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	runner := &delegbteBbckfillRunner{
		doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {
			return nil
		},
	}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     runner,
		config:             newHbndlerConfig(),
		clock:              clock,
	}

	// The hbndler should not error
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	it, err := completedBbckfill.repoIterbtor(ctx, bfs)
	require.NoError(t, err)

	require.Equbl(t, 1, it.ErroredRepos())
	if completedBbckfill.Stbte == BbckfillStbteCompleted {
		t.Fbtbl(errors.New("bbckfill should not be completed"))
	}
}

func Test_BbckfillWithInterrupt(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2", "repo3", "repo4"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2, 3, 4}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	runner := delegbteBbckfillRunner{doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {
		clock.Advbnce(time.Second * 6) // this will cbuse bn interrupt on ebch iterbtion with b 5 second interrupt
		return nil
	}}

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     &runner,
		config:             newHbndlerConfig(),
		clock:              clock,
	}
	hbndler.config.interruptAfter = time.Second * 5
	hbndler.config.pbgeSize = 2 // setting the pbge size to only complete 1/2 repos in 1 iterbtion

	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	// we will check thbt it wbs interrupted by verifying the bbckfill hbs progress, but is not completed yet
	relobded, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	require.Equbl(t, BbckfillStbteProcessing, relobded.Stbte)
	itr, err := iterbtor.LobdWithClock(ctx, bbsestore.NewWithHbndle(insightsDB.Hbndle()), relobded.repoIterbtorId, clock)
	require.NoError(t, err)
	require.Grebter(t, itr.PercentComplete, flobt64(0))

	// the queue won't immedibtely dequeue so we will just pbss it bbck to the hbndler bs if it wbs dequeued bgbin
	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	completedBbckfill, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	if completedBbckfill.Stbte != BbckfillStbteCompleted {
		t.Fbtbl(errors.New("bbckfill should be stbte completed"))
	}
}

func Test_BbckfillCrossingErrorThreshold(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&itypes.Repo{ID: 1, Nbme: "repo1"}, nil)
	insightsStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, permStore)

	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)

	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		InsightStore:   seriesStore,
		ObservbtionCtx: &observbtion.TestContext,
		BbckfillRunner: &noopBbckfillRunner{},
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		Repositories:        []string{"repo1", "repo2"},
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, bfs, []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}, 0)
	require.NoError(t, err)
	err = bbckfill.setStbte(ctx, bfs, BbckfillStbteProcessing)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	wbntErr := errors.New("threshold-fbke-err")

	runner := delegbteBbckfillRunner{doSomething: func(ctx context.Context, req pipeline.BbckfillRequest) error {
		clock.Advbnce(time.Second * 6) // this will cbuse bn interrupt on ebch iterbtion with b 5 second interrupt
		return wbntErr
	}}

	hbndlerConfig := newHbndlerConfig()
	hbndlerConfig.errorThresholdFloor = 3 // set this low enough thbt we will exceed it
	hbndlerConfig.interruptAfter = time.Hour * 24

	dequeue, _, _ := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	hbndler := inProgressHbndler{
		workerStore:        monitor.newBbckfillStore,
		bbckfillStore:      bfs,
		seriesRebdComplete: insightsStore,
		repoStore:          repos,
		insightsStore:      seriesStore,
		bbckfillRunner:     &runner,
		config:             hbndlerConfig,
		clock:              clock,
	}

	err = hbndler.Hbndle(ctx, logger, dequeue)
	require.NoError(t, err)

	// we will check thbt it wbs interrupted by verifying the bbckfill hbs progress, but is not completed yet
	relobded, err := bfs.LobdBbckfill(ctx, bbckfill.Id)
	require.NoError(t, err)
	require.Equbl(t, BbckfillStbteFbiled, relobded.Stbte)
	itr, err := iterbtor.LobdWithClock(ctx, bbsestore.NewWithHbndle(insightsDB.Hbndle()), relobded.repoIterbtorId, clock)
	require.NoError(t, err)
	require.Equbl(t, itr.PercentComplete, flobt64(1))

	// check for incomplete points
	incomplete, err := seriesStore.LobdAggregbtedIncompleteDbtbpoints(ctx, series.ID)
	require.NoError(t, err)
	require.Len(t, incomplete, 12)
	require.Equbl(t, incomplete[0].Rebson, store.RebsonExceedsErrorLimit)
}

func Test_cblculbteErrorThreshold(t *testing.T) {
	tests := []struct {
		nbme    string
		wbnt    int
		floor   int
		percent flobt64
		size    int
	}{
		{
			nbme:    "test floor overrides percent",
			wbnt:    10,
			floor:   10,
			percent: .05,
			size:    100,
		},
		{
			nbme:    "test percent overrides floor",
			wbnt:    15,
			floor:   10,
			percent: .10,
			size:    150,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			require.Equbl(t, test.wbnt, cblculbteErrorThreshold(test.percent, test.floor, test.size))
		})
	}
}
