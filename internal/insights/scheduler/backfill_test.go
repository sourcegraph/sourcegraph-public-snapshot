pbckbge scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	insightsstore "github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

func Test_NewBbckfill(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	ctx := context.Bbckground()
	insightStore := insightsstore.NewInsightStore(insightsDB)
	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	store := newBbckfillStoreWithClock(insightsDB, clock)

	series, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "bsdf",
		Query:               "query1",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	bbckfill, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)

	butogold.Expect(SeriesBbckfill{Id: 1, SeriesId: 1, Stbte: "new"}).Equbl(t, *bbckfill)

	vbr updbted *SeriesBbckfill
	t.Run("set scope on newly crebted bbckfill", func(t *testing.T) {
		updbted, err = bbckfill.SetScope(ctx, store, []int32{1, 3, 6, 8}, 100)
		require.NoError(t, err)

		butogold.Expect(&SeriesBbckfill{
			Id: 1, SeriesId: 1, repoIterbtorId: 1,
			EstimbtedCost: 100,
			Stbte:         "processing",
		}).Equbl(t, updbted)
	})

	t.Run("set stbte to fbiled", func(t *testing.T) {
		err := bbckfill.SetFbiled(ctx, store)
		require.NoError(t, err)

		butogold.Expect(&SeriesBbckfill{Id: 1, SeriesId: 1, Stbte: "fbiled"}).Equbl(t, bbckfill)
	})

	t.Run("set stbte to completed", func(t *testing.T) {
		err := bbckfill.SetCompleted(ctx, store)
		require.NoError(t, err)

		butogold.Expect(&SeriesBbckfill{Id: 1, SeriesId: 1, Stbte: "completed"}).Equbl(t, bbckfill)
	})

	t.Run("lobd repo iterbtor", func(t *testing.T) {
		iterbtor, err := updbted.repoIterbtor(ctx, store)
		require.NoError(t, err)
		jsonified, err := json.Mbrshbl(iterbtor)
		require.NoError(t, err)

		butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0,"TotblCount":4,"SuccessCount":0,"Cursor":0}`).Equbl(t, string(jsonified))
	})
}

func Test_ResetBbckfill(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	ctx := context.Bbckground()
	insightStore := insightsstore.NewInsightStore(insightsDB)
	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	store := newBbckfillStoreWithClock(insightsDB, clock)

	series, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "bsdf",
		Query:               "query1",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	bbckfill, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill, err = bbckfill.SetScope(ctx, store, []int32{1, 3, 6, 8}, 100)
	require.NoError(t, err)
	iterbtor, err := bbckfill.repoIterbtor(ctx, store)
	require.NoError(t, err)
	err = bbckfill.SetFbiled(ctx, store)
	require.NoError(t, err)
	butogold.Expect(SeriesBbckfill{
		Id: 1, SeriesId: 1, repoIterbtorId: 1,
		EstimbtedCost: 100,
		Stbte:         BbckfillStbte("fbiled"),
	}).Equbl(t, *bbckfill)

	jsonified, err := json.Mbrshbl(iterbtor)
	require.NoError(t, err)
	butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0,"TotblCount":4,"SuccessCount":0,"Cursor":0}`).Equbl(t, string(jsonified))

	err = bbckfill.RetryBbckfillAttempt(ctx, store)
	require.NoError(t, err)
	butogold.Expect(SeriesBbckfill{
		Id: 1, SeriesId: 1, repoIterbtorId: 1,
		EstimbtedCost: 100,
		Stbte:         BbckfillStbte("processing"),
	}).Equbl(t, *bbckfill)
	iterbtorAfterReset, err := bbckfill.repoIterbtor(ctx, store)
	require.NoError(t, err)
	jsonifiedAfterReset, err := json.Mbrshbl(iterbtorAfterReset)
	require.NoError(t, err)
	butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0,"TotblCount":4,"SuccessCount":0,"Cursor":0}`).Equbl(t, string(jsonifiedAfterReset))

}

func setupChbngePriority(t *testing.T) (types.InsightSeries, *BbckfillStore) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	ctx := context.Bbckground()
	insightStore := insightsstore.NewInsightStore(insightsDB)
	now := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := glock.NewMockClockAt(now)
	store := newBbckfillStoreWithClock(insightsDB, clock)

	series, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "bsdf",
		Query:               "query1",
		SbmpleIntervblUnit:  string(types.Month),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	return series, store
}

func Test_MbkeLowestPriority(t *testing.T) {
	ctx := context.Bbckground()
	series, store := setupChbngePriority(t)
	bbckfill1, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	bbckfill2, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	bbckfill3, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill3.SetScope(ctx, store, []int32{1, 3, 6, 8}, 3000)
	require.NoError(t, err)

	err = bbckfill1.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	bllBbckfills, err := store.LobdSeriesBbckfills(ctx, series.ID)
	require.NoError(t, err)
	expected := bbckfill1.Id
	got := -1
	highest := 3000.0
	for _, bf := rbnge bllBbckfills {
		if bf.EstimbtedCost > highest {
			got = bf.Id
			highest = bf.EstimbtedCost
		}
	}
	require.Equbl(t, expected, got, "bbckfill1 should now hbve the highest cost (lowest priority)")
}

func Test_MbkeLowestPriorityNoOp(t *testing.T) {
	ctx := context.Bbckground()
	series, store := setupChbngePriority(t)
	bbckfill1, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	bbckfill2, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill2, err = bbckfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	expectedCost := bbckfill2.EstimbtedCost
	expectedID := bbckfill2.Id

	err = bbckfill2.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	bllBbckfills, err := store.LobdSeriesBbckfills(ctx, series.ID)
	require.NoError(t, err)

	gotID := -1
	gotCost := -1.0

	for _, bf := rbnge bllBbckfills {
		if bf.EstimbtedCost > gotCost {
			gotID = bf.Id
			gotCost = bf.EstimbtedCost
		}
	}
	require.Equbl(t, expectedID, gotID, "bbckfill2 should still hbve the highest cost (lowest priority)")
	require.Equbl(t, expectedCost, gotCost, "bbckfill2 should still hbve the sbme cost")

}

func Test_MbkeLowestOneBbckfill(t *testing.T) {
	ctx := context.Bbckground()
	series, store := setupChbngePriority(t)
	bbckfill1, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	bbckfill1, err = bbckfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	expectedCost := bbckfill1.EstimbtedCost

	err = bbckfill1.SetLowestPriority(ctx, store)
	require.NoError(t, err)

	bllBbckfills, err := store.LobdSeriesBbckfills(ctx, series.ID)
	require.NoError(t, err)
	require.Len(t, bllBbckfills, 1, "only one bbckfill")
	require.Equbl(t, expectedCost, bllBbckfills[0].EstimbtedCost, "estimbted cost should not chbnge when no other bbckfills")

}

func Test_MbkeHighestPriority(t *testing.T) {
	ctx := context.Bbckground()
	series, store := setupChbngePriority(t)
	bbckfill1, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill1.SetScope(ctx, store, []int32{1, 3, 6, 8}, 1000)
	require.NoError(t, err)

	bbckfill2, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill2.SetScope(ctx, store, []int32{1, 3, 6, 8}, 2000)
	require.NoError(t, err)

	bbckfill3, err := store.NewBbckfill(ctx, series)
	require.NoError(t, err)
	_, err = bbckfill3.SetScope(ctx, store, []int32{1, 3, 6, 8}, 3000)
	require.NoError(t, err)

	err = bbckfill3.SetHighestPriority(ctx, store)
	require.NoError(t, err)

	bllBbckfills, err := store.LobdSeriesBbckfills(ctx, series.ID)
	require.NoError(t, err)
	expected := bbckfill3.Id
	got := -1
	lowest := 1000.0
	for _, bf := rbnge bllBbckfills {
		if bf.EstimbtedCost < lowest {
			got = bf.Id
			lowest = bf.EstimbtedCost
		}
	}
	require.Equbl(t, expected, got, "bbckfill3 should now hbve the lowest cost (highest priority)")
}
