package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	insightsstore "github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
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

	autogold.Expect(SeriesBackfill{Id: 1, SeriesId: 1, State: "new"}).Equal(t, *backfill)

	var updated *SeriesBackfill
	t.Run("set scope on newly created backfill", func(t *testing.T) {
		updated, err = backfill.SetScope(ctx, store, []int32{1, 3, 6, 8}, 100)
		require.NoError(t, err)

		autogold.Expect(&SeriesBackfill{
			Id: 1, SeriesId: 1, repoIteratorId: 1,
			EstimatedCost: 100,
			State:         "processing",
		}).Equal(t, updated)
	})

	t.Run("set state to failed", func(t *testing.T) {
		err := backfill.SetFailed(ctx, store)
		require.NoError(t, err)

		autogold.Expect(&SeriesBackfill{Id: 1, SeriesId: 1, State: "failed"}).Equal(t, backfill)
	})

	t.Run("set state to completed", func(t *testing.T) {
		err := backfill.SetCompleted(ctx, store)
		require.NoError(t, err)

		autogold.Expect(&SeriesBackfill{Id: 1, SeriesId: 1, State: "completed"}).Equal(t, backfill)
	})

	t.Run("load repo iterator", func(t *testing.T) {
		iterator, err := updated.repoIterator(ctx, store)
		require.NoError(t, err)
		jsonified, err := json.Marshal(iterator)
		require.NoError(t, err)

		autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":4,"SuccessCount":0,"Cursor":0}`).Equal(t, string(jsonified))
	})
}

func Test_ResetBackfill(t *testing.T) {
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
	backfill, err = backfill.SetScope(ctx, store, []int32{1, 3, 6, 8}, 100)
	require.NoError(t, err)
	iterator, err := backfill.repoIterator(ctx, store)
	require.NoError(t, err)
	err = backfill.SetFailed(ctx, store)
	require.NoError(t, err)
	autogold.Expect(SeriesBackfill{
		Id: 1, SeriesId: 1, repoIteratorId: 1,
		EstimatedCost: 100,
		State:         BackfillState("failed"),
	}).Equal(t, *backfill)

	jsonified, err := json.Marshal(iterator)
	require.NoError(t, err)
	autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":4,"SuccessCount":0,"Cursor":0}`).Equal(t, string(jsonified))

	err = backfill.RetryBackfillAttempt(ctx, store)
	require.NoError(t, err)
	autogold.Expect(SeriesBackfill{
		Id: 1, SeriesId: 1, repoIteratorId: 1,
		EstimatedCost: 100,
		State:         BackfillState("processing"),
	}).Equal(t, *backfill)
	iteratorAfterReset, err := backfill.repoIterator(ctx, store)
	require.NoError(t, err)
	jsonifiedAfterReset, err := json.Marshal(iteratorAfterReset)
	require.NoError(t, err)
	autogold.Expect(`{"Id":1,"CreatedAt":"2021-01-01T00:00:00Z","StartedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDuration":0,"PercentComplete":0,"TotalCount":4,"SuccessCount":0,"Cursor":0}`).Equal(t, string(jsonifiedAfterReset))

}

func setupChangePriority(t *testing.T) (types.InsightSeries, *BackfillStore) {
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
	return series, store
}

func Test_MakeLowestPriority(t *testing.T) {
	ctx := context.Background()
	series, store := setupChangePriority(t)
	backfill1, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	backfill2, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	backfill3, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill3.SetScope(ctx, store, []int32{1, 3, 6, 8}, 3000)
	require.NoError(t, err)

	err = backfill1.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	allBackfills, err := store.LoadSeriesBackfills(ctx, series.ID)
	require.NoError(t, err)
	expected := backfill1.Id
	got := -1
	highest := 3000.0
	for _, bf := range allBackfills {
		if bf.EstimatedCost > highest {
			got = bf.Id
			highest = bf.EstimatedCost
		}
	}
	require.Equal(t, expected, got, "backfill1 should now have the highest cost (lowest priority)")
}

func Test_MakeLowestPriorityNoOp(t *testing.T) {
	ctx := context.Background()
	series, store := setupChangePriority(t)
	backfill1, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	backfill2, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	backfill2, err = backfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	expectedCost := backfill2.EstimatedCost
	expectedID := backfill2.Id

	err = backfill2.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	allBackfills, err := store.LoadSeriesBackfills(ctx, series.ID)
	require.NoError(t, err)

	gotID := -1
	gotCost := -1.0

	for _, bf := range allBackfills {
		if bf.EstimatedCost > gotCost {
			gotID = bf.Id
			gotCost = bf.EstimatedCost
		}
	}
	require.Equal(t, expectedID, gotID, "backfill2 should still have the highest cost (lowest priority)")
	require.Equal(t, expectedCost, gotCost, "backfill2 should still have the same cost")

}

func Test_MakeLowestOneBackfill(t *testing.T) {
	ctx := context.Background()
	series, store := setupChangePriority(t)
	backfill1, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	backfill1, err = backfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	expectedCost := backfill1.EstimatedCost

	err = backfill1.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	allBackfills, err := store.LoadSeriesBackfills(ctx, series.ID)
	require.NoError(t, err)
	require.Len(t, allBackfills, 1, "only one backfill")
	require.Equal(t, expectedCost, allBackfills[0].EstimatedCost, "estimated cost should not change when no other backfills")

}

func Test_MakeHighestPriority(t *testing.T) {
	ctx := context.Background()
	series, store := setupChangePriority(t)
	backfill1, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	backfill2, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	backfill3, err := store.NewBackfill(ctx, series)
	require.NoError(t, err)
	_, err = backfill3.SetScope(ctx, store, []int32{1, 3, 6, 8}, 3000)
	require.NoError(t, err)

	err = backfill3.SetHighestPriority(ctx, store)
	require.NoError(t, err)

	allBackfills, err := store.LoadSeriesBackfills(ctx, series.ID)
	require.NoError(t, err)
	expected := backfill3.Id
	got := -1
	lowest := 1000.0
	for _, bf := range allBackfills {
		if bf.EstimatedCost < lowest {
			got = bf.Id
			lowest = bf.EstimatedCost
		}
	}
	require.Equal(t, expected, got, "backfill3 should now have the lowest cost (highest priority)")
}
