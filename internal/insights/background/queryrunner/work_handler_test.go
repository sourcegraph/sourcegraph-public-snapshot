package queryrunner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	store2 "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/hexops/autogold/v2"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

func TestGetSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	metadataStore := store.NewInsightStore(insightsDB)
	metadataStore.Now = func() time.Time {
		return now
	}
	ctx := context.Background()

	workHandler := workHandler{
		metadadataStore: metadataStore,
		mu:              sync.RWMutex{},
		seriesCache:     make(map[string]*types.InsightSeries),
	}

	t.Run("series definition does not exist", func(t *testing.T) {
		_, err := workHandler.getSeries(ctx, "seriesshouldnotexist")
		if err == nil {
			t.Fatal("expected error from getSeries")
		}
		autogold.Expect("workHandler.getSeries: insight definition not found for series_id: seriesshouldnotexist").Equal(t, err.Error())
	})

	t.Run("series definition does exist", func(t *testing.T) {
		series, err := metadataStore.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "arealseries",
			Query:                      "query1",
			CreatedAt:                  now,
			OldestHistoricalAt:         now,
			LastRecordedAt:             now,
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			BackfillQueuedAt:           now,
			Enabled:                    true,
			Repositories:               nil,
			SampleIntervalUnit:         string(types.Month),
			SampleIntervalValue:        1,
			GeneratedFromCaptureGroups: false,
			JustInTime:                 false,
			GenerationMethod:           types.Search,
		})
		if err != nil {
			t.Error(err)
		}
		got, err := workHandler.getSeries(ctx, series.SeriesID)
		if err != nil {
			t.Fatal("unexpected error from getseries")
		}
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func Test_HandleWithTerminalError(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(t))
	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	metadataStore := store.NewInsightStore(insightsDB)
	metadataStore.Now = func() time.Time {
		return now
	}
	ctx := context.Background()

	setUp := func(t *testing.T, seriesId string) types.InsightSeries {
		series, err := metadataStore.CreateSeries(ctx, types.InsightSeries{
			SeriesID:            seriesId,
			Query:               "findme",
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 5,
			GenerationMethod:    types.Search,
		})
		if err != nil {
			t.Fatal(err)
		}
		return series
	}

	tss := store.New(insightsDB, store.NewInsightPermissionStore(postgres))
	fakeErr := errors.New("fake err")

	handlers := make(map[types.GenerationMethod]InsightsHandler)
	handlers[types.Search] = func(ctx context.Context, job *SearchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		return nil, fakeErr
	}
	workerStore := CreateDBWorkerStore(observation.TestContextTB(t), basestore.NewWithHandle(postgres.Handle()))

	queueIt := func(t *testing.T, previousFailures int, series types.InsightSeries) *Job {
		job := &Job{
			SearchJob: SearchJob{
				SeriesID:    series.SeriesID,
				SearchQuery: "findme",
				RecordTime:  nil, // set nil to emulate a global query
				PersistMode: string(store.RecordMode),
			},
			State:    "queued",
			Cost:     10,
			Priority: 10,
		}
		id, err := EnqueueJob(ctx, basestore.NewWithHandle(workerStore.Handle()), job)
		if err != nil {
			t.Fatal(err)
		}
		job.ID = id
		err = basestore.NewWithHandle(workerStore.Handle()).Exec(ctx, sqlf.Sprintf("update insights_query_runner_jobs set num_failures = %s where id = %s", previousFailures, job.ID))
		if err != nil {
			t.Fatal(err)
		}
		job.NumFailures = int32(previousFailures)
		return job
	}

	handler := &workHandler{
		insightsStore:   tss,
		baseWorkerStore: workerStore,
		metadadataStore: metadataStore,
		limiter:         ratelimit.NewInstrumentedLimiter("asdf", rate.NewLimiter(10, 5)),
		logger:          logger,
		mu:              sync.RWMutex{},
		seriesCache:     make(map[string]*types.InsightSeries),
		searchHandlers:  handlers,
	}

	t.Run("ensure max errors produces incomplete point entry", func(t *testing.T) {
		series := setUp(t, "terminal")
		job := queueIt(t, 9, series)
		err := handler.Handle(ctx, logger, job)
		require.ErrorIs(t, err, fakeErr)
		incompletes, err := tss.LoadAggregatedIncompleteDatapoints(ctx, series.ID)
		if err != nil {
			t.Fatal(err)
		}
		require.Len(t, incompletes, 1)
		_, err = workerStore.MarkComplete(ctx, job.ID, store2.MarkFinalOptions{})
		require.NoError(t, err)
	})
	t.Run("ensure less than max errors does not produce an incomplete point entry", func(t *testing.T) {
		series := setUp(t, "willretry")
		job := queueIt(t, 7, series)
		err := handler.Handle(ctx, logger, job)
		require.ErrorIs(t, err, fakeErr)
		incompletes, err := tss.LoadAggregatedIncompleteDatapoints(ctx, series.ID)
		if err != nil {
			t.Fatal(err)
		}
		require.Empty(t, incompletes)
		_, err = workerStore.MarkComplete(ctx, job.ID, store2.MarkFinalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
	})
}
