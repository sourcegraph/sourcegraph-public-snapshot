package queryrunner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/hexops/autogold"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGetSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
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
		autogold.Want("series definition does not exist", "workHandler.getSeries: insight definition not found for series_id: seriesshouldnotexist").Equal(t, err.Error())
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
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

}

func Test_updateSeriesRecordingTimes(t *testing.T) {
	buildRecordingTimes := func(numPoints int, interval timeseries.TimeInterval, now time.Time) []time.Time {
		frames := timeseries.BuildFrames(numPoints, interval, now)
		return timeseries.GetRecordingTimesFromFrames(frames)
	}

	now := time.Date(2022, 4, 4, 0, 0, 0, 0, time.UTC)
	then := now.AddDate(0, -1, 0)

	testCases := []struct {
		name                string
		recordingTimes      []time.Time
		newTime             time.Time
		wantAdd, wantDelete autogold.Value
	}{
		{
			"empty doesn't break",
			[]time.Time{},
			now,
			autogold.Want("add", []time.Time{now}),
			autogold.Want("nothing to delete", []time.Time{}),
		},
		{
			"less than 12 recordings just adds new",
			buildRecordingTimes(5, timeseries.TimeInterval{types.Month, 5}, then),
			now,
			autogold.Want("add new value", []time.Time{now}),
			autogold.Want("nothing to delete", []time.Time{}),
		},
		{
			"more than 12 recordings but all within the past year",
			buildRecordingTimes(14, timeseries.TimeInterval{types.Day, 5}, then),
			now,
			autogold.Want("add new value", []time.Time{now}),
			autogold.Want("nothing to delete", []time.Time{}),
		},
		{
			"more than 12 recordings over multiple years delete oldest",
			buildRecordingTimes(12, timeseries.TimeInterval{types.Year, 1}, then),
			now,
			autogold.Want("add new value", []time.Time{now}),
			autogold.Want("delete oldest value", []time.Time{then.AddDate(-11, 0, 0)}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAdd, gotDelete := updateSeriesRecordingTimes(tc.recordingTimes, tc.newTime)
			tc.wantAdd.Equal(t, gotAdd)
			tc.wantDelete.Equal(t, gotDelete)
		})
	}
}
