package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/insights/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
		logger := logtest.Scoped(t)
		clock := func() time.Time { return now }
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
		postgres := database.NewDB(logger, dbtest.NewDB(t))
		resolver := newWithClock(insightsDB, postgres, clock)
		insightStore := store.NewInsightStore(insightsDB)

		view := types.InsightView{
			Title:            "title1",
			Description:      "desc1",
			PresentationType: types.Line,
		}
		insightSeries := types.InsightSeries{
			SeriesID:            "1234567",
			Query:               "query1",
			CreatedAt:           now,
			OldestHistoricalAt:  now,
			LastRecordedAt:      now,
			NextRecordingAfter:  now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		}
		var err error
		view, err = insightStore.CreateView(ctx, view, []store.InsightViewGrant{store.GlobalGrant()})
		require.NoError(t, err)
		insightSeries, err = insightStore.CreateSeries(ctx, insightSeries)
		require.NoError(t, err)
		insightStore.AttachSeriesToView(ctx, insightSeries, view, types.InsightViewSeriesMetadata{
			Label:  "",
			Stroke: "",
		})

		insightMetadataStore := store.NewMockInsightMetadataStore()

		resolver.insightMetadataStore = insightMetadataStore

		// Create the insightview connection resolver and query series.
		conn, err := resolver.InsightViews(ctx, &graphqlbackend.InsightViewQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		nodes, err := conn.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var series [][]graphqlbackend.InsightSeriesResolver
		for _, node := range nodes {
			s, _ := node.DataSeries(ctx)
			series = append(series, s)
		}
		return ctx, series
	}

	t.Run("Points", func(t *testing.T) {
		ctx, insights := testSetup(t)
		autogold.Expect(1).Equal(t, len(insights))

		autogold.Expect(1).Equal(t, len(insights[0]))

		// Issue a query against the actual DB.
		points, err := insights[0][0].Points(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]graphqlbackend.InsightsDataPointResolver{}).Equal(t, points)
	})
}

func fakeStatusGetter(status *queryrunner.JobsStatus, err error) GetSeriesQueueStatusFunc {
	return func(ctx context.Context, seriesID string) (*queryrunner.JobsStatus, error) {
		return status, err
	}
}

func fakeBackfillGetter(backfills []scheduler.SeriesBackfill, err error) GetSeriesBackfillsFunc {
	return func(ctx context.Context, seriesID int) ([]scheduler.SeriesBackfill, error) {
		return backfills, err
	}
}

func fakeIncompleteGetter() GetIncompleteDatapointsFunc {
	return func(ctx context.Context, seriesID int) ([]store.IncompleteDatapoint, error) {
		return nil, nil
	}
}

func TestInsightSeriesStatusResolver_IsLoadingData(t *testing.T) {
	type isLoadingTestCase struct {
		name         string
		backfills    []scheduler.SeriesBackfill
		backfillsErr error
		queueStatus  queryrunner.JobsStatus
		queueErr     error
		series       types.InsightViewSeries
		want         autogold.Value
	}

	recentTime := time.Date(2020, time.April, 1, 1, 0, 0, 0, time.UTC)

	cases := []isLoadingTestCase{
		{
			name:      "completed backvillv2",
			backfills: []scheduler.SeriesBackfill{{State: scheduler.BackfillStateCompleted}},
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:false error:"),
		},
		{
			name:      "completed backfillv1",
			backfills: []scheduler.SeriesBackfill{},
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:false error:"),
		},
		{
			name:      "new backfillv2",
			backfills: []scheduler.SeriesBackfill{{State: scheduler.BackfillStateNew}},
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:true error:"),
		},
		{
			name:      "in process backfillv2",
			backfills: []scheduler.SeriesBackfill{{State: scheduler.BackfillStateProcessing}},
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:true error:"),
		},
		{
			name:      "in process backfillv1",
			backfills: []scheduler.SeriesBackfill{},
			queueStatus: queryrunner.JobsStatus{
				Queued:     10,
				Processing: 2,
				Errored:    1,
			},
			series: types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:   autogold.Expect("loading:true error:"),
		},
		{
			name:      "failed backfillv2",
			backfills: []scheduler.SeriesBackfill{{State: scheduler.BackfillStateFailed}},
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:false error:"),
		},
		{
			name:      "failed backfillv1",
			backfills: []scheduler.SeriesBackfill{},
			queueStatus: queryrunner.JobsStatus{
				Failed: 10,
			},
			series: types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:   autogold.Expect("loading:false error:"),
		},
		{
			name:      "completed but snapshotting backfillv2",
			backfills: []scheduler.SeriesBackfill{{State: scheduler.BackfillStateCompleted}},
			queueStatus: queryrunner.JobsStatus{
				Queued: 1,
			},
			series: types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:   autogold.Expect("loading:true error:"),
		},
		{
			name:         "error loading backfill",
			backfills:    []scheduler.SeriesBackfill{},
			backfillsErr: errors.New("backfill error"),
			series:       types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:         autogold.Expect("loading:false error:LoadSeriesBackfills: backfill error"),
		},
		{
			name:      "error loading queue status",
			backfills: []scheduler.SeriesBackfill{},
			queueErr:  errors.New("error loading queue status"),
			series:    types.InsightViewSeries{BackfillQueuedAt: &recentTime},
			want:      autogold.Expect("loading:false error:QueryJobsStatus: error loading queue status"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statusGetter := fakeStatusGetter(&tc.queueStatus, tc.queueErr)
			backfillGetter := fakeBackfillGetter(tc.backfills, tc.backfillsErr)
			statusResolver := newStatusResolver(statusGetter, backfillGetter, fakeIncompleteGetter(), tc.series)
			loading, err := statusResolver.IsLoadingData(context.Background())
			var loadingResult bool
			if loading != nil {
				loadingResult = *loading
			}
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			tc.want.Equal(t, fmt.Sprintf("loading:%t error:%s", loadingResult, errMsg))
		})
	}
}

func TestInsightStatusResolver_IncompleteDatapoints(t *testing.T) {
	// Setup the GraphQL resolver.
	ctx := actor.WithInternalActor(context.Background())
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(t))
	insightStore := store.NewInsightStore(insightsDB)
	tss := store.New(insightsDB, store.NewInsightPermissionStore(postgres))

	base := baseInsightResolver{
		insightStore:    insightStore,
		timeSeriesStore: tss,
		insightsDB:      insightsDB,
		postgresDB:      postgres,
	}

	series, err := insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "asdf",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	require.NoError(t, err)

	repo := 5
	addFakeIncomplete := func(in time.Time) {
		err = tss.AddIncompleteDatapoint(ctx, store.AddIncompleteDatapointInput{
			SeriesID: series.ID,
			RepoID:   &repo,
			Reason:   store.ReasonTimeout,
			Time:     in,
		})
		require.NoError(t, err)
	}

	resolver := NewStatusResolver(&base, types.InsightViewSeries{InsightSeriesID: series.ID})

	addFakeIncomplete(now)
	addFakeIncomplete(now)
	addFakeIncomplete(now.AddDate(0, 0, 1))

	stringify := func(input []graphqlbackend.IncompleteDatapointAlert) (res []string) {
		for _, in := range input {
			res = append(res, in.Time().String())
		}
		return res
	}

	t.Run("as timeout", func(t *testing.T) {
		got, err := resolver.IncompleteDatapoints(ctx)
		require.NoError(t, err)
		autogold.Expect([]string{"2020-01-01 00:00:00 +0000 UTC", "2020-01-02 00:00:00 +0000 UTC"}).Equal(t, stringify(got))
	})
}

func Test_NumSamplesFiltering(t *testing.T) {
	// Setup the GraphQL resolver.
	ctx := actor.WithInternalActor(context.Background())
	// now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(t))
	insightStore := store.NewInsightStore(insightsDB)
	tss := store.New(insightsDB, store.NewInsightPermissionStore(postgres))

	series, err := insightStore.CreateSeries(ctx, types.InsightSeries{
		ID:                  0,
		SeriesID:            "asdf",
		Query:               "asdf",
		SampleIntervalUnit:  string(types.Month),
		SampleIntervalValue: 1,
	})
	require.NoError(t, err)

	repo := "repo1"
	repoId := api.RepoID(1)

	times := []types.RecordingTime{
		{Timestamp: time.Date(2023, 2, 2, 16, 25, 40, 0, time.UTC), Snapshot: true},
		{Timestamp: time.Date(2023, 2, 2, 16, 25, 36, 0, time.UTC), Snapshot: false},
		{Timestamp: time.Date(2023, 1, 30, 18, 12, 39, 0, time.UTC), Snapshot: false},
		{Timestamp: time.Date(2023, 1, 25, 15, 34, 23, 0, time.UTC), Snapshot: false},
	}

	err = tss.RecordSeriesPointsAndRecordingTimes(ctx, []store.RecordSeriesPointArgs{
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[0].Timestamp,
				Value:    10,
			},
			RepoName:    &repo,
			RepoID:      &repoId,
			PersistMode: store.SnapshotMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[1].Timestamp,
				Value:    10,
			},
			RepoName:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[2].Timestamp,
				Value:    10,
			},
			RepoName:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
		{
			SeriesID: series.SeriesID,
			Point: store.SeriesPoint{
				SeriesID: series.SeriesID,
				Time:     times[3].Timestamp,
				Value:    10,
			},
			RepoName:    &repo,
			RepoID:      &repoId,
			PersistMode: store.RecordMode,
		},
	}, types.InsightSeriesRecordingTimes{InsightSeriesID: series.ID, RecordingTimes: times})
	require.NoError(t, err)

	base := baseInsightResolver{
		insightStore:    insightStore,
		timeSeriesStore: tss,
		insightsDB:      insightsDB,
		postgresDB:      postgres,
	}

	tests := []struct {
		name       string
		numSamples int32
	}{
		{
			name:       "one",
			numSamples: 1,
		},
		{
			name:       "two",
			numSamples: 2,
		},
		{
			name:       "three",
			numSamples: 3,
		},
		{
			name:       "four",
			numSamples: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			points, err := fetchSeries(ctx, types.InsightViewSeries{SeriesID: series.SeriesID, InsightSeriesID: series.ID}, types.InsightViewFilters{}, types.SeriesDisplayOptions{NumSamples: &test.numSamples}, &base)
			require.NoError(t, err)

			assert.Equal(t, int(test.numSamples), len(points))
			t.Log(points)
			for i := range points {
				assert.Equal(t, times[len(points)-i-1].Timestamp, points[i].Time)
			}
		})
	}
}
