package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log/logtest"
)

func Test_MonitorStartsAndStops(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		ObservationCtx: &observation.TestContext,
		CostAnalyzer:   priority.NewQueryAnalyzer(),
	}
	routines := NewBackgroundJobMonitor(ctx, config).Routines()
	goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

func TestScheduler_InitialBackfill(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	repos := dbmocks.NewMockRepoStore()
	insightsStore := store.NewInsightStore(insightsDB)
	config := JobMonitorConfig{
		InsightsDB:     insightsDB,
		RepoStore:      repos,
		ObservationCtx: &observation.TestContext,
		CostAnalyzer:   priority.NewQueryAnalyzer(),
	}
	monitor := NewBackgroundJobMonitor(ctx, config)

	series, err := insightsStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series1",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	require.NoError(t, err)

	scheduler := NewScheduler(insightsDB)
	backfill, err := scheduler.InitialBackfill(ctx, series)
	require.NoError(t, err)

	dequeue, found, err := monitor.newBackfillStore.Dequeue(ctx, "test", nil)
	require.NoError(t, err)
	if !found {
		t.Fatal(errors.New("no queued record found"))
	}
	require.Equal(t, backfill.Id, dequeue.backfillId)
}
