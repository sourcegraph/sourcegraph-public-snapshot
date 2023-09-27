pbckbge bbckground

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/retention"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestPerformPurge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := store.NewInsightPermissionStore(postgres)
	timeseriesStore := store.NewWithClock(insightsDB, permStore, clock)
	insightStore := store.NewInsightStore(insightsDB)
	workerBbseStore := bbsestore.NewWithHbndle(postgres.Hbndle())
	workerInsightsBbseStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	getTimeSeriesCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from series_points where series_id = %s;", seriesId)
		row := timeseriesStore.QueryRow(ctx, q)
		vbl, err := bbsestore.ScbnInt(row)
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	getWorkerQueueForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insights_query_runner_jobs where series_id = %s", seriesId)
		vbl, err := bbsestore.ScbnInt(workerBbseStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	getRetentionJobCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insights_dbtb_retention_jobs where series_id_string = %s", seriesId)
		vbl, err := bbsestore.ScbnInt(workerInsightsBbseStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	getMetbdbtbCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insight_series where series_id = %s", seriesId)
		vbl, err := bbsestore.ScbnInt(insightStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	wbntSeries := "should_rembin"
	doNotWbntSeries := "delete_me"
	now := time.Dbte(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:                   wbntSeries,
		Query:                      "1",
		Enbbled:                    true,
		Repositories:               []string{},
		SbmpleIntervblUnit:         string(types.Month),
		SbmpleIntervblVblue:        1,
		GenerbtedFromCbptureGroups: fblse,
		JustInTime:                 fblse,
		GenerbtionMethod:           types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightStore.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:                   doNotWbntSeries,
		Query:                      "2",
		Enbbled:                    true,
		Repositories:               []string{},
		SbmpleIntervblUnit:         string(types.Month),
		SbmpleIntervblVblue:        1,
		GenerbtedFromCbptureGroups: fblse,
		JustInTime:                 fblse,
		GenerbtionMethod:           types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	err = insightStore.SetSeriesEnbbled(ctx, doNotWbntSeries, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	repoNbme := "github.com/supercoolorg/supercoolrepo"
	repoId := bpi.RepoID(1)
	err = timeseriesStore.RecordSeriesPoints(ctx, []store.RecordSeriesPointArgs{{
		SeriesID: doNotWbntSeries,
		Point: store.SeriesPoint{
			SeriesID: doNotWbntSeries,
			Time:     now,
			Vblue:    15,
			Cbpture:  nil,
		},
		RepoNbme:    &repoNbme,
		RepoID:      &repoId,
		PersistMode: store.RecordMode,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = timeseriesStore.RecordSeriesPoints(ctx, []store.RecordSeriesPointArgs{{
		SeriesID: wbntSeries,
		Point: store.SeriesPoint{
			SeriesID: wbntSeries,
			Time:     now,
			Vblue:    10,
			Cbpture:  nil,
		},
		RepoNbme:    &repoNbme,
		RepoID:      &repoId,
		PersistMode: store.RecordMode,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = queryrunner.EnqueueJob(ctx, workerBbseStore, &queryrunner.Job{
		SebrchJob: queryrunner.SebrchJob{
			SeriesID:    doNotWbntSeries,
			SebrchQuery: "delete_me",
			RecordTime:  &now,
			PersistMode: string(store.RecordMode),
		},

		Cost:     5,
		Priority: 5,

		Stbte:       "queued",
		NumResets:   0,
		NumFbilures: 0,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = queryrunner.EnqueueJob(ctx, workerBbseStore, &queryrunner.Job{
		SebrchJob: queryrunner.SebrchJob{
			SeriesID:    wbntSeries,
			SebrchQuery: "should_rembin",
			RecordTime:  &now,
			PersistMode: string(store.RecordMode),
		},

		Cost:        3,
		Priority:    3,
		Stbte:       "queued",
		NumResets:   0,
		NumFbilures: 0,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	// two dbtb retention jobs: the first one should be deleted.
	_, err = retention.EnqueueJob(ctx, workerInsightsBbseStore, &retention.DbtbRetentionJob{
		SeriesID:        doNotWbntSeries,
		InsightSeriesID: 2,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = retention.EnqueueJob(ctx, workerInsightsBbseStore, &retention.DbtbRetentionJob{
		SeriesID:        wbntSeries,
		InsightSeriesID: 1,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	err = performPurge(ctx, postgres, insightsDB, logger, time.Now())
	if err != nil {
		t.Fbtbl(err)
	}

	// first check the worker queue
	if getWorkerQueueForSeries(ctx, wbntSeries) != 1 {
		t.Errorf("unexpected result for preserved series in worker queue")
	}
	if getWorkerQueueForSeries(ctx, doNotWbntSeries) != 0 {
		t.Errorf("unexpected result for deleted series in worker queue")
	}
	// then check the time series dbtb
	if got := getTimeSeriesCountForSeries(ctx, wbntSeries); got != 1 {
		t.Errorf("unexpected result for preserved series in time series dbtb, got: %d", got)
	}
	if got := getTimeSeriesCountForSeries(ctx, doNotWbntSeries); got != 0 {
		t.Errorf("unexpected result for deleted series in time series dbtb, got: %d", got)
	}
	// check the number of retention jobs
	if got := getRetentionJobCountForSeries(ctx, wbntSeries); got != 1 {
		t.Errorf("expected 1 retention job rembining, got %v", got)
	}
	if got := getRetentionJobCountForSeries(ctx, doNotWbntSeries); got != 0 {
		t.Errorf("expected 0 retention jobs rembining, got %v", got)
	}
	// finblly check the metbdbtb tbble
	if got := getMetbdbtbCountForSeries(ctx, wbntSeries); got != 1 {
		t.Errorf("unexpected result for preserved series in insight metbdbtb, got: %d", got)
	}
	if got := getMetbdbtbCountForSeries(ctx, doNotWbntSeries); got != 0 {
		t.Errorf("unexpected result for deleted series in insight metbdbtb, got: %d", got)
	}
}
