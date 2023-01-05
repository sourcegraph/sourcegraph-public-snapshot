package retention

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func Test_selectOldestRecordingTimeBeforeMax(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mainDB := database.NewDB(logger, dbtest.NewDB(logger, t))

	insightStore := store.NewInsightStore(insightsDB)
	seriesStore := store.New(insightsDB, store.NewInsightPermissionStore(mainDB))

	// create a series with id 1 to attach to recording times
	setupSeries(ctx, insightStore, t)

	recordingTimes := types.InsightSeriesRecordingTimes{InsightSeriesID: 1}
	newTime := time.Now().Truncate(time.Hour)
	var expectedOldestTimestamp time.Time
	for i := 1; i <= 12; i++ {
		newTime = newTime.Add(time.Hour)
		recordingTimes.RecordingTimes = append(recordingTimes.RecordingTimes, types.RecordingTime{
			Snapshot: false, Timestamp: newTime,
		})
		if i == 7 {
			expectedOldestTimestamp = newTime
		}
	}
	if err := seriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{recordingTimes}); err != nil {
		t.Fatal(err)
	}

	got, err := selectOldestRecordingTimeBeforeMax(ctx, seriesStore, 1, 6)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.String() != expectedOldestTimestamp.String() {
		t.Errorf("expected timestamp %v got %v", expectedOldestTimestamp, got)
	}
}

func Test_archiveOldSeriesPoints(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	mainDB := database.NewDB(logger, dbtest.NewDB(logger, t))

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
	oldestTimestamp, err := selectOldestRecordingTimeBeforeMax(ctx, seriesStore, 1, sampleSize)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveOldSeriesPoints(ctx, seriesStore, seriesID, *oldestTimestamp); err != nil {
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
	mainDB := database.NewDB(logger, dbtest.NewDB(logger, t))

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
	oldestTimestamp, err := selectOldestRecordingTimeBeforeMax(ctx, seriesStore, 1, sampleSize)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveOldRecordingTimes(ctx, seriesStore, 1, *oldestTimestamp); err != nil {
		t.Fatal(err)
	}

	got, err := seriesStore.GetInsightSeriesRecordingTimes(ctx, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RecordingTimes) != sampleSize {
		t.Errorf("expected 4 recording times left, got %d", len(got.RecordingTimes))
	}
}

func TestHandle_ErrorDuringTransaction(t *testing.T) {
	// This tests that if we error at any point during sql execution we will roll back, and we will not lose any data.

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
