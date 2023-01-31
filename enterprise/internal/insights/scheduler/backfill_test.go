package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	insightsstore "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func Test_NewBackfill(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	ctx := context.Background()
	insightStore := insightsstore.NewInsightStore(insightsDB)
	now := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	store := newBackfillStoreWithClock(insightsDB, clock)

	series, err := insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "asdf",
		Query:               "query1",
		SampleIntervalUnit:  string(types.Month),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}

	backfill, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)

	autogold.Want("backfill loaded successfully", SeriesBackfill{Id: 1, SeriesId: 1, State: "new"}).Equal(t, *backfill)

	var updated *SeriesBackfill
	t.Run("set scope on newly created backfill", func(t *testing.T) {
		updated, err = backfill.SetScope(ctx, store, []int32{1, 3, 6, 8}, 100)
		require.NoError(t, err)

		autogold.Want("set scope on newly created backfill", &SeriesBackfill{
			Id: 1, SeriesId: 1, repoIteratorId: 1,
			EstimatedCost: 100,
			State:         "processing",
		}).Equal(t, updated)
	})

	t.Run("set state to failed", func(t *testing.T) {
		err := backfill.SetFailed(ctx, store)
		require.NoError(t, err)

		autogold.Want("set state to failed", &SeriesBackfill{Id: 1, SeriesId: 1, State: "failed"}).Equal(t, backfill)
	})

	t.Run("set state to completed", func(t *testing.T) {
		err := backfill.SetCompleted(ctx, store)
		require.NoError(t, err)

		autogold.Want("set state to completed", &SeriesBackfill{Id: 1, SeriesId: 1, State: "completed"}).Equal(t, backfill)
	})

	t.Run("load repo iterator", func(t *testing.T) {
		iterator, err := updated.repoIterator(ctx, store)
		require.NoError(t, err)
		jsonified, err := json.Marshal(iterator)
		require.NoError(t, err)

		autogold.Want("load repo iterator", `{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":4,"SuccessCount":0,"Cursor":0}`).Equal(t, string(jsonified))
	})
}
