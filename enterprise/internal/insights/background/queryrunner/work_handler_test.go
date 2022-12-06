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
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
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
