package background

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/retention"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestPerformPurge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(t))
	permStore := store.NewInsightPermissionStore(postgres)
	timeseriesStore := store.NewWithClock(insightsDB, permStore, clock)
	insightStore := store.NewInsightStore(insightsDB)
	workerBaseStore := basestore.NewWithHandle(postgres.Handle())
	workerInsightsBaseStore := basestore.NewWithHandle(insightsDB.Handle())

	getTimeSeriesCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from series_points where series_id = %s;", seriesId)
		row := timeseriesStore.QueryRow(ctx, q)
		val, err := basestore.ScanInt(row)
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	getWorkerQueueForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insights_query_runner_jobs where series_id = %s", seriesId)
		val, err := basestore.ScanInt(workerBaseStore.QueryRow(ctx, q))
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	getRetentionJobCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insights_data_retention_jobs where series_id_string = %s", seriesId)
		val, err := basestore.ScanInt(workerInsightsBaseStore.QueryRow(ctx, q))
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	getMetadataCountForSeries := func(ctx context.Context, seriesId string) int {
		q := sqlf.Sprintf("select count(*) from insight_series where series_id = %s", seriesId)
		val, err := basestore.ScanInt(insightStore.QueryRow(ctx, q))
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	wantSeries := "should_remain"
	doNotWantSeries := "delete_me"
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:                   wantSeries,
		Query:                      "1",
		Enabled:                    true,
		Repositories:               []string{},
		SampleIntervalUnit:         string(types.Month),
		SampleIntervalValue:        1,
		GeneratedFromCaptureGroups: false,
		JustInTime:                 false,
		GenerationMethod:           types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:                   doNotWantSeries,
		Query:                      "2",
		Enabled:                    true,
		Repositories:               []string{},
		SampleIntervalUnit:         string(types.Month),
		SampleIntervalValue:        1,
		GeneratedFromCaptureGroups: false,
		JustInTime:                 false,
		GenerationMethod:           types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = insightStore.SetSeriesEnabled(ctx, doNotWantSeries, false)
	if err != nil {
		t.Fatal(err)
	}
	repoName := "github.com/supercoolorg/supercoolrepo"
	repoId := api.RepoID(1)
	err = timeseriesStore.RecordSeriesPoints(ctx, []store.RecordSeriesPointArgs{{
		SeriesID: doNotWantSeries,
		Point: store.SeriesPoint{
			SeriesID: doNotWantSeries,
			Time:     now,
			Value:    15,
			Capture:  nil,
		},
		RepoName:    &repoName,
		RepoID:      &repoId,
		PersistMode: store.RecordMode,
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = timeseriesStore.RecordSeriesPoints(ctx, []store.RecordSeriesPointArgs{{
		SeriesID: wantSeries,
		Point: store.SeriesPoint{
			SeriesID: wantSeries,
			Time:     now,
			Value:    10,
			Capture:  nil,
		},
		RepoName:    &repoName,
		RepoID:      &repoId,
		PersistMode: store.RecordMode,
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = queryrunner.EnqueueJob(ctx, workerBaseStore, &queryrunner.Job{
		SearchJob: queryrunner.SearchJob{
			SeriesID:    doNotWantSeries,
			SearchQuery: "delete_me",
			RecordTime:  &now,
			PersistMode: string(store.RecordMode),
		},

		Cost:     5,
		Priority: 5,

		State:       "queued",
		NumResets:   0,
		NumFailures: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queryrunner.EnqueueJob(ctx, workerBaseStore, &queryrunner.Job{
		SearchJob: queryrunner.SearchJob{
			SeriesID:    wantSeries,
			SearchQuery: "should_remain",
			RecordTime:  &now,
			PersistMode: string(store.RecordMode),
		},

		Cost:        3,
		Priority:    3,
		State:       "queued",
		NumResets:   0,
		NumFailures: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	// two data retention jobs: the first one should be deleted.
	_, err = retention.EnqueueJob(ctx, workerInsightsBaseStore, &retention.DataRetentionJob{
		SeriesID:        doNotWantSeries,
		InsightSeriesID: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = retention.EnqueueJob(ctx, workerInsightsBaseStore, &retention.DataRetentionJob{
		SeriesID:        wantSeries,
		InsightSeriesID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = performPurge(ctx, postgres, insightsDB, logger, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	// first check the worker queue
	if getWorkerQueueForSeries(ctx, wantSeries) != 1 {
		t.Errorf("unexpected result for preserved series in worker queue")
	}
	if getWorkerQueueForSeries(ctx, doNotWantSeries) != 0 {
		t.Errorf("unexpected result for deleted series in worker queue")
	}
	// then check the time series data
	if got := getTimeSeriesCountForSeries(ctx, wantSeries); got != 1 {
		t.Errorf("unexpected result for preserved series in time series data, got: %d", got)
	}
	if got := getTimeSeriesCountForSeries(ctx, doNotWantSeries); got != 0 {
		t.Errorf("unexpected result for deleted series in time series data, got: %d", got)
	}
	// check the number of retention jobs
	if got := getRetentionJobCountForSeries(ctx, wantSeries); got != 1 {
		t.Errorf("expected 1 retention job remaining, got %v", got)
	}
	if got := getRetentionJobCountForSeries(ctx, doNotWantSeries); got != 0 {
		t.Errorf("expected 0 retention jobs remaining, got %v", got)
	}
	// finally check the metadata table
	if got := getMetadataCountForSeries(ctx, wantSeries); got != 1 {
		t.Errorf("unexpected result for preserved series in insight metadata, got: %d", got)
	}
	if got := getMetadataCountForSeries(ctx, doNotWantSeries); got != 0 {
		t.Errorf("unexpected result for deleted series in insight metadata, got: %d", got)
	}
}
