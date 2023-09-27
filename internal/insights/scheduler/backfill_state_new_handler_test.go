pbckbge scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Test_MovesBbckfillFromNewToProcessing(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefbultReturn([]*itypes.Repo{{ID: 1, Nbme: "repo1"}, {ID: 2, Nbme: "repo2"}}, nil)
	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)
	insightsStore := store.NewInsightStore(insightsDB)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	seriesStore := store.New(insightsDB, permStore)
	repoQueryExecutor := NewMockRepoQueryExecutor()
	repoQueryExecutor.ExecuteRepoListFunc.SetDefbultReturn(nil, errors.New("repo query executor should not be cblled"))

	config := JobMonitorConfig{
		InsightsDB:        insightsDB,
		RepoStore:         repos,
		ObservbtionCtx:    &observbtion.TestContext,
		CostAnblyzer:      priority.NewQueryAnblyzer(),
		InsightStore:      seriesStore,
		RepoQueryExecutor: repoQueryExecutor,
	}
	vbr err error
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

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	newDequeue, _, err := monitor.newBbckfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	hbndler := newBbckfillHbndler{
		workerStore:     monitor.newBbckfillStore,
		bbckfillStore:   bfs,
		seriesRebder:    store.NewInsightStore(insightsDB),
		repoIterbtor:    discovery.NewSeriesRepoIterbtor(nil, repos, repoQueryExecutor),
		costAnblyzer:    *config.CostAnblyzer,
		timeseriesStore: seriesStore,
	}
	err = hbndler.Hbndle(ctx, logger, newDequeue)
	require.NoError(t, err)

	_, dupFound, err := monitor.newBbckfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if dupFound {
		t.Fbtbl(errors.New("found record thbt should not be visible to the new bbckfill store"))
	}

	// now ensure the in progress hbndler _cbn_ pick it up
	inProgressDequeue, inProgressFound, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !inProgressFound {
		t.Fbtbl(errors.New("no queued record found"))
	}
	require.Equbl(t, bbckfill.Id, inProgressDequeue.bbckfillId)

	recordingTimes, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, series.ID, store.SeriesPointsOpts{})
	require.NoError(t, err)
	if len(recordingTimes.RecordingTimes) == 0 {
		t.Fbtbl(errors.New("recording times should hbve been sbved bfter success"))
	}
}

func Test_MovesBbckfillFromNewToProcessing_ScopedInsight(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefbultReturn([]*itypes.Repo{}, errors.New("the repo store should not be cblled"))
	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	bfs := newBbckfillStoreWithClock(insightsDB, clock)
	insightsStore := store.NewInsightStore(insightsDB)
	permStore := store.NewInsightPermissionStore(dbmocks.NewMockDB())
	seriesStore := store.New(insightsDB, permStore)
	repoQueryExecutor := NewMockRepoQueryExecutor()
	repoQueryExecutor.ExecuteRepoListFunc.SetDefbultReturn([]itypes.MinimblRepo{{Nbme: "sourcegrbph/sourcegrbph", ID: 1}}, nil)

	config := JobMonitorConfig{
		InsightsDB:        insightsDB,
		RepoStore:         repos,
		ObservbtionCtx:    &observbtion.TestContext,
		CostAnblyzer:      priority.NewQueryAnblyzer(),
		InsightStore:      seriesStore,
		RepoQueryExecutor: repoQueryExecutor,
	}
	vbr err error
	monitor := NewBbckgroundJobMonitor(ctx, config)

	repoCriterib := "repo:sourcegrbph"
	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		RepositoryCriterib:  &repoCriterib,
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	bbckfill, err := bfs.NewBbckfill(ctx, series)
	require.NoError(t, err)

	err = enqueueBbckfill(ctx, bfs.Hbndle(), bbckfill)
	require.NoError(t, err)

	newDequeue, _, err := monitor.newBbckfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	hbndler := newBbckfillHbndler{
		workerStore:     monitor.newBbckfillStore,
		bbckfillStore:   bfs,
		seriesRebder:    store.NewInsightStore(insightsDB),
		repoIterbtor:    discovery.NewSeriesRepoIterbtor(nil, repos, repoQueryExecutor),
		costAnblyzer:    *config.CostAnblyzer,
		timeseriesStore: seriesStore,
	}
	err = hbndler.Hbndle(ctx, logger, newDequeue)
	require.NoError(t, err)

	_, dupFound, err := monitor.newBbckfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if dupFound {
		t.Fbtbl(errors.New("found record thbt should not be visible to the new bbckfill store"))
	}

	// now ensure the in progress hbndler _cbn_ pick it up
	inProgressDequeue, inProgressFound, err := monitor.inProgressStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !inProgressFound {
		t.Fbtbl(errors.New("no queued record found"))
	}
	require.Equbl(t, bbckfill.Id, inProgressDequeue.bbckfillId)

	recordingTimes, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, series.ID, store.SeriesPointsOpts{})
	require.NoError(t, err)
	if len(recordingTimes.RecordingTimes) == 0 {
		t.Fbtbl(errors.New("recording times should hbve been sbved bfter success"))
	}
}
