package retention

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Test_archiveOldSeriesPoints(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mainDB := database.NewDB(logger, dbtest.NewDB(t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mainDB))

	// create a series with id 1 and name 'series1' to attach to recording times
	setupSeries(ctx, insightStore, t)
	seriesID := "series1"

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncate(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = append(recordingTimes.RecordingTimes, types.RecordingTime{
			Snapshot: false, Timestamp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fatal(err)
	}

	// Insert some series points
	_, err := insightsDB.ExecContext(context.Background(), `
SELECT setseed(0.5);
INSERT INTO series_points(
    time,
	series_id,
    value
)
SELECT recording_time,
    'series1',
    random()*80 - 40
	FROM insight_series_recording_times WHERE insight_series_id = 1;
`)
	if err != nil {
		t.Fatal(err)
	}

	sampleSize := 8
	oldestTimestamp, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, sampleSize-1, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveOldSeriesPoints(ctx, seriesStore, seriesID, oldestTimestamp); err != nil {
		t.Fatal(err)
	}

	got, err := seriesStore.SeriesPoints(ctx, store.SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != sampleSize {
		t.Errorf("expected 8 series points, got %d", len(got))
	}
}

func Test_archiveOldRecordingTimes(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mainDB := database.NewDB(logger, dbtest.NewDB(t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mainDB))

	// create a series with id 1 to attach to recording times
	setupSeries(ctx, insightStore, t)

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncate(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = append(recordingTimes.RecordingTimes, types.RecordingTime{
			Snapshot: false, Timestamp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fatal(err)
	}

	sampleSize := 4
	oldestTimestamp, err := seriesStore.GetOffsetNRecordingTime(ctx, 1, sampleSize-1, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveOldRecordingTimes(ctx, seriesStore, 1, oldestTimestamp); err != nil {
		t.Fatal(err)
	}

	got, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, 1, store.SeriesPointsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RecordingTimes) != sampleSize {
		t.Errorf("expected 4 recording times left, got %d", len(got.RecordingTimes))
	}
}

func TestHandle_ErrorDuringTransaction(t *testing.T) {
	// This tests that if we error at any point during sql execution we will roll back, and we will not lose any data.
	ctx := context.Background()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mainDB := database.NewDB(logger, dbtest.NewDB(t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mainDB))

	baseWorkerStore := basestore.NewWithHandle(insightsDB.Handle())
	workerStore := CreateDBWorkerStore(observation.TestContextTB(t), baseWorkerStore)

	boolTrue := true
	conf.Get().ExperimentalFeatures.InsightsDataRetention = &boolTrue
	conf.Get().InsightsMaximumSampleSize = 2
	t.Cleanup(func() {
		conf.Get().InsightsMaximumSampleSize = 0
		conf.Get().ExperimentalFeatures.InsightsDataRetention = nil
	})

	handler := &dataRetentionHandler{
		baseWorkerStore: workerStore,
		insightsStore:   seriesStore,
	}

	setupSeries(ctx, insightStore, t)

	// setup recording times
	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncate(time.Hour)
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = append(recordingTimes.RecordingTimes, types.RecordingTime{
			Snapshot: false, Timestamp: newTime,
		})
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fatal(err)
	}

	// drop a table. create chaos
	_, err := insightsDB.ExecContext(context.Background(), `
DROP TABLE IF EXISTS series_points
`)
	if err != nil {
		t.Fatal(err)
	}

	job := &DataRetentionJob{SeriesID: "series1", InsightSeriesID: 1}
	id, err := EnqueueJob(ctx, baseWorkerStore, job)
	if err != nil {
		t.Fatal(err)
	}
	job.ID = id

	err = handler.Handle(ctx, logger, job)
	if err == nil {
		t.Fatal("expected error got nil")
	}

	got, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, 1, store.SeriesPointsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RecordingTimes) != 12 {
		t.Errorf("expected 12 recording times still remaining after rollback, got %d", len(got.RecordingTimes))
	}
}

func setupSeries(ctx context.Context, tx *store.InsightStore, t *testing.T) {
	now := time.Now()
	series := types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query-1",
		OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
		LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		Enabled:            true,
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
	}
	got, err := tx.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 {
		t.Errorf("expected first series to have id 1")
	}
}

func Test_GetSampleSize(t *testing.T) {
	logger := logtest.Scoped(t)

	t.Run("not configured", func(t *testing.T) {
		conf.Mock(&conf.Unified{})
		assert.Equal(t, 30, getMaximumSampleSize(logger))
	})

	t.Run("exceeds max value", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{InsightsMaximumSampleSize: 100}})
		assert.Equal(t, 90, getMaximumSampleSize(logger))
	})

	t.Run("negative value", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{InsightsMaximumSampleSize: -40}})
		assert.Equal(t, 30, getMaximumSampleSize(logger))
	})

	t.Run("matches config", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{InsightsMaximumSampleSize: 50}})
		assert.Equal(t, 50, getMaximumSampleSize(logger))
	})
}
