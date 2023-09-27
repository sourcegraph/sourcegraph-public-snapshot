pbckbge scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log/logtest"
)

func Test_MonitorStbrtsAndStops(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second*1)
	defer cbncel()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		ObservbtionCtx: &observbtion.TestContext,
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	routines := NewBbckgroundJobMonitor(ctx, config).Routines()
	goroutine.MonitorBbckgroundRoutines(ctx, routines...)
}

func TestScheduler_InitiblBbckfill(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	insightsStore := store.NewInsightStore(insightsDB)
	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		ObservbtionCtx: &observbtion.TestContext,
		CostAnblyzer:   priority.NewQueryAnblyzer(),
	}
	monitor := NewBbckgroundJobMonitor(ctx, config)

	series, err := insightsStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "bsdf",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(insightsDB)
	bbckfill, err := scheduler.InitiblBbckfill(ctx, series)
	require.NoError(t, err)

	dequeue, found, err := monitor.newBbckfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !found {
		t.Fbtbl(errors.New("no queued record found"))
	}
	require.Equbl(t, bbckfill.Id, dequeue.bbckfillId)
}
