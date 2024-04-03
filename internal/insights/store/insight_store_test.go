package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestGet(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	groupByRepo := "repo"

	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
									VALUES (1, 'test title', 'test description', 'unique-1', false),
									       (2, 'test title 2', 'test description 2', 'unique-2', true)`)
	if err != nil {
		t.Fatal(err)
	}

	// assign some global grants just so the test can immediately fetch the created views
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_grants (insight_view_id, global)
									VALUES (1, true),
									       (2, true)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, last_snapshot_at, next_snapshot_after, deleted_at, generation_method, group_by, repository_criteria)
                            VALUES ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'search', null,'repo:a'),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'search', 'repo', null),
									('series-id-3-deleted', 'query-3', $1, $1, $1, $1, $1, $1, $1, 'search', null, 'repo:*');`, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
									VALUES (1, 1, 'label1', 'color1'),
											(1, 2, 'label2', 'color2'),
											(2, 2, 'second-label-2', 'second-color-2'),
											(2, 3, 'label3', 'color-2');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("test get all", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label1",
				LineColor:            "color1",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label2",
				LineColor:            "color2",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               2,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "second-label-2",
				LineColor:            "second-color-2",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             true,
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})

	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueIDs: []string{"unique-1"}})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label1",
				LineColor:            "color1",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label2",
				LineColor:            "color2",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: "unique-1"})
		if err != nil {
			t.Fatal(err)
		}
		sampleIntervalUnit := "MONTH"
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label1",
				LineColor:            "color1",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				SampleIntervalValue:  1,
				SampleIntervalUnit:   sampleIntervalUnit,
				Label:                "label2",
				LineColor:            "color2",
				PresentationType:     types.Line,
				GenerationMethod:     types.Search,
				IsFrozen:             false,
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
}

func TestGetAll(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	groupByRepo := "repo"
	ctx := context.Background()

	store := NewInsightStore(insightsDB)

	// First test the method on an empty database.
	t.Run("test empty database", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff([]types.InsightViewSeries{}, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})

	// Set up some insight views to test pagination and permissions.
	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id)
	VALUES (1, 'user cannot view', '', 'a'),
		   (2, 'user can view 1', '', 'd'),
		   (3, 'user can view 2', '', 'e'),
		   (4, 'user cannot view 2', '', 'f'),
		   (5, 'user can view 3', '', 'b')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (id, series_id, query, created_at, oldest_historical_at, last_recorded_at,
		next_recording_after, last_snapshot_at, next_snapshot_after, deleted_at, generation_method, group_by, repository_criteria)
		VALUES  (1, 'series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'search', null, 'repo:a'),
				(2, 'series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'search', 'repo', null)`, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
	VALUES  (1, 1, 'label1-1', 'color'),
			(2, 1, 'label2-1', 'color'),
			(2, 2, 'label2-2', 'color'),
			(3, 1, 'label3-1', 'color'),
			(4, 1, 'label4-1', 'color'),
			(4, 2, 'label4-2', 'color'),
			(5, 1, 'label5-1', 'color'),
			(5, 2, 'label5-2', 'color');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_grants (insight_view_id, global)
	VALUES (2, true), (3, true)`)
	if err != nil {
		t.Fatal(err)
	}

	// Attach one of the insights to a dashboard to test insight permission via dashboard permissions.
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (id, title) VALUES (1, 'dashboard 1');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id) VALUES (1, 5)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (dashboard_id, global) VALUES (1, true)`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("all results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               3,
				UniqueID:             "e",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 2",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label3-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("first result", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAll(ctx, InsightQueryArgs{Limit: 1})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("second result", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Limit: 1, After: "b"})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("last 2 results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{After: "b"})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               3,
				UniqueID:             "e",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 2",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label3-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("find by title results", func(*testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Find: "view 3"})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("find by series label results", func(*testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Find: "label5-1"})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 3",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label5-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("exclude insight ids from results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{ExcludeIDs: []string{"b", "e"}})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-1",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user can view 1",
				Description:          "",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("returns expected number of samples", func(t *testing.T) {
		// Set the series_num_samples value
		numSamples := int32(50)
		view, err := store.UpdateView(ctx, types.InsightView{
			UniqueID:         "d",
			PresentationType: types.Line, // setting for null constraint
			SeriesNumSamples: &numSamples,
		})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(&numSamples, view.SeriesNumSamples); diff != "" {
			t.Errorf("unexpected insight view series num samples want/got: %s", diff)
		}

		series, err := store.GetAll(ctx, InsightQueryArgs{UniqueIDs: []string{"d"}})
		if err != nil {
			t.Fatal(err)
		}
		// we're only testing the number of samples in this test cases
		for _, s := range series {
			if diff := cmp.Diff(&numSamples, s.SeriesNumSamples); diff != "" {
				t.Errorf("unexpected insight view series num samples want/got: %s", diff)
			}
		}
	})
}

func TestGetAllOnDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	groupByRepo := "repo"

	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'test title', 'test description', 'unique-1'),
									       (2, 'test title 2', 'test description 2', 'unique-2'),
										   (3, 'test title 3', 'test description 3', 'unique-3'),
										   (4, 'test title 4', 'test description 4', 'unique-4')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, last_snapshot_at, next_snapshot_after, deleted_at, generation_method, group_by, repository_criteria)
                            VALUES  ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'search', null, 'repo:a'),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'search', 'repo', null),
									('series-id-3-deleted', 'query-3', $1, $1, $1, $1, $1, $1, $1, 'search', null, null);`, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
									VALUES  (1, 1, 'label1-1', 'color1'),
											(2, 2, 'label2-2', 'color2'),
											(3, 1, 'label3-1', 'color3'),
											(4, 2, 'label4-2', 'color4');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (id, title) VALUES  (1, 'dashboard 1');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
									VALUES  (1, 2),
											(1, 1),
											(1, 4),
											(1, 3);`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("test get all on dashboard", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDashboard(ctx, InsightsOnDashboardQueryArgs{DashboardID: 1})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               2,
				DashboardViewID:      1,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color2",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               1,
				DashboardViewID:      2,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label1-1",
				LineColor:            "color1",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
			{
				ViewID:               4,
				DashboardViewID:      3,
				UniqueID:             "unique-4",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 4",
				Description:          "test description 4",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label4-2",
				LineColor:            "color4",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               3,
				DashboardViewID:      4,
				UniqueID:             "unique-3",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title 3",
				Description:          "test description 3",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label3-1",
				LineColor:            "color3",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("test get first 2 on dashboard", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDashboard(ctx, InsightsOnDashboardQueryArgs{DashboardID: 1, Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               2,
				DashboardViewID:      1,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label2-2",
				LineColor:            "color2",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               1,
				DashboardViewID:      2,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label1-1",
				LineColor:            "color1",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
	t.Run("test get after 2 on dashboard", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDashboard(ctx, InsightsOnDashboardQueryArgs{DashboardID: 1, After: "2"})
		if err != nil {
			t.Fatal(err)
		}
		series1RepoCriteria := "repo:a"
		want := []types.InsightViewSeries{
			{
				ViewID:               4,
				DashboardViewID:      3,
				UniqueID:             "unique-4",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 4",
				Description:          "test description 4",
				Query:                "query-2",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label4-2",
				LineColor:            "color4",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				GroupBy:              &groupByRepo,
				SupportsAugmentation: true,
			},
			{
				ViewID:               3,
				DashboardViewID:      4,
				UniqueID:             "unique-3",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title 3",
				Description:          "test description 3",
				Query:                "query-1",
				CreatedAt:            now,
				OldestHistoricalAt:   now,
				LastRecordedAt:       now,
				NextRecordingAfter:   now,
				LastSnapshotAt:       now,
				NextSnapshotAfter:    now,
				Label:                "label3-1",
				LineColor:            "color3",
				SampleIntervalUnit:   "MONTH",
				SampleIntervalValue:  1,
				PresentationType:     types.PresentationType("LINE"),
				GenerationMethod:     types.GenerationMethod("search"),
				SupportsAugmentation: true,
				RepositoryCriteria:   &series1RepoCriteria,
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected insight view series want/got: %s", diff)
		}
	})
}

func TestCreateSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 5, 1, 1, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	groupByRepo := "repo"

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	ctx := context.Background()

	t.Run("test create series", func(t *testing.T) {
		repoCriteria := "repo:a"
		series := types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
			GenerationMethod:   types.Search,
			GroupBy:            &groupByRepo,
			RepositoryCriteria: &repoCriteria,
		}

		got, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		want := types.InsightSeries{
			ID:                   1,
			SeriesID:             "unique-1",
			Query:                "query-1",
			OldestHistoricalAt:   now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			CreatedAt:            now,
			Enabled:              true,
			SampleIntervalUnit:   string(types.Month),
			GenerationMethod:     types.Search,
			GroupBy:              &groupByRepo,
			SupportsAugmentation: true,
			RepositoryCriteria:   &repoCriteria,
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result from create insight series (want/got): %s", diff)
		}
	})
	t.Run("test create and get capture groups series", func(t *testing.T) {
		sampleIntervalUnit := "MONTH"
		repoCriteria := "repo:a"
		_, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "capture-group-1",
			Query:                      "well hello there",
			Enabled:                    true,
			SampleIntervalUnit:         sampleIntervalUnit,
			SampleIntervalValue:        0,
			OldestHistoricalAt:         now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:             now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			CreatedAt:                  now,
			GeneratedFromCaptureGroups: true,
			GenerationMethod:           types.Search,
			RepositoryCriteria:         &repoCriteria,
		})
		if err != nil {
			return
		}

		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			SeriesID: "capture-group-1",
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) < 1 {
			t.Fatal(err)
		}
		got[0].ID = 1 // normalizing this for test determinism

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestCreateView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test create view", func(t *testing.T) {
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
			Filters: types.InsightViewFilters{
				SearchContexts: []string{"@dev/mycontext"},
			},
		}

		got, err := store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}

		want := types.InsightView{
			ID:               1,
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
			Filters: types.InsightViewFilters{
				SearchContexts: []string{"@dev/mycontext"},
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result from create insight view (want/got): %s", diff)
		}
	})
}

func TestCreateGetView_WithGrants(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CreateView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentationType: types.Line,
	}, []InsightViewGrant{UserGrant(1), OrgGrant(5)})
	if err != nil {
		t.Fatal(err)
	}
	series, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CreatedAt:          now,
		OldestHistoricalAt: now,
		LastRecordedAt:     now,
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		BackfillQueuedAt:   now,
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "label1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("user 1 can see this view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for user 1 insight views")
		}
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})

	t.Run("user 2 cannot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{2}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for user 2 insight views")
		}
	})

	t.Run("org 1 cannot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{3}, OrgIDs: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for org 1 insight views")
		}
	})
	t.Run("org 5 can see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{3}, OrgIDs: []int{5}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for org 5 insight views")
		}
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("no users or orgs provided should only return global", func(t *testing.T) {
		uniqueID := "globalonly"
		view, err := store.CreateView(ctx, types.InsightView{
			Title:            "global only",
			Description:      "global only",
			UniqueID:         uniqueID,
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		series, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "globalseries",
			Query:              "global",
			CreatedAt:          now,
			OldestHistoricalAt: now,
			LastRecordedAt:     now,
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			BackfillQueuedAt:   now,
			SampleIntervalUnit: string(types.Month),
			GenerationMethod:   types.Search,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
			Label:  "label2",
			Stroke: "red",
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("unexpected count for global only insights")
		}
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestUpdateView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test update view", func(t *testing.T) {
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		got, err := store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(types.InsightView{
			ID: 1, Title: "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}).Equal(t, got)

		include, exclude := "include repos", "exclude repos"
		got, err = store.UpdateView(ctx, types.InsightView{
			Title:    "new title",
			UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: &include,
				ExcludeRepoRegex: &exclude,
				SearchContexts:   []string{"@dev/mycontext"},
			},
			PresentationType: types.Line,
		})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(types.InsightView{
			ID: 1, Title: "new title", UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: valast.Addr("include repos").(*string),
				ExcludeRepoRegex: valast.Addr("exclude repos").(*string),
				SearchContexts:   []string{"@dev/mycontext"},
			},
			PresentationType: "LINE",
		}).Equal(t, got)
	})
}

func TestUpdateViewSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	groupByRepo := "repo"
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test update view series", func(t *testing.T) {
		view, err := store.CreateView(ctx, types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		series, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			Enabled:            true,
			SampleIntervalUnit: string(types.Month),
			GenerationMethod:   types.Search,
			GroupBy:            &groupByRepo,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
			Label:  "label",
			Stroke: "blue",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = store.UpdateViewSeries(ctx, series.SeriesID, view.ID, types.InsightViewSeriesMetadata{
			Label:  "new label",
			Stroke: "orange",
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect("new label").Equal(t, got[0].Label)
		autogold.Expect("orange").Equal(t, got[0].LineColor)
	})
}

func TestDeleteView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CreateView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentationType: types.Line,
	}, []InsightViewGrant{GlobalGrant()})
	if err != nil {
		t.Fatal(err)
	}
	series, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CreatedAt:          now,
		OldestHistoricalAt: now,
		LastRecordedAt:     now,
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		BackfillQueuedAt:   now,
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "label1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("delete view and check length", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) < 1 {
			t.Errorf("expected results before deleting view")
		}
		err = store.DeleteViewByUniqueID(ctx, uniqueID)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("expected results after deleting view")
		}
	})
}

func TestAttachSeriesView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test attach and fetch", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricalAt:  now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
			GenerationMethod:    types.Search,
		}
		series, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		view, err = store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		metadata := types.InsightViewSeriesMetadata{
			Label:  "my label",
			Stroke: "my stroke",
		}
		err = store.AttachSeriesToView(ctx, series, view, metadata)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{{
			ViewID:               1,
			UniqueID:             view.UniqueID,
			InsightSeriesID:      series.ID,
			SeriesID:             series.SeriesID,
			Title:                view.Title,
			Description:          view.Description,
			Query:                series.Query,
			CreatedAt:            series.CreatedAt,
			OldestHistoricalAt:   series.OldestHistoricalAt,
			LastRecordedAt:       series.LastRecordedAt,
			NextRecordingAfter:   series.NextRecordingAfter,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			SampleIntervalValue:  1,
			SampleIntervalUnit:   sampleIntervalUnit,
			Label:                "my label",
			LineColor:            "my stroke",
			PresentationType:     types.Line,
			GenerationMethod:     types.Search,
			SupportsAugmentation: true,
		}}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after attaching series to view (want/got): %s", diff)
		}
	})
}

func TestRemoveSeriesFromView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test remove series from view", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricalAt:  now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
			GenerationMethod:    types.Search,
		}
		series, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentationType: types.Line,
		}
		view, err = store.CreateView(ctx, view, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		metadata := types.InsightViewSeriesMetadata{
			Label:  "my label",
			Stroke: "my stroke",
		}
		err = store.AttachSeriesToView(ctx, series, view, metadata)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		sampleIntervalUnit := "MONTH"
		want := []types.InsightViewSeries{{
			ViewID:               1,
			UniqueID:             view.UniqueID,
			InsightSeriesID:      series.ID,
			SeriesID:             series.SeriesID,
			Title:                view.Title,
			Description:          view.Description,
			Query:                series.Query,
			CreatedAt:            series.CreatedAt,
			OldestHistoricalAt:   series.OldestHistoricalAt,
			LastRecordedAt:       series.LastRecordedAt,
			NextRecordingAfter:   series.NextRecordingAfter,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			SampleIntervalValue:  1,
			SampleIntervalUnit:   sampleIntervalUnit,
			Label:                "my label",
			LineColor:            "my stroke",
			PresentationType:     types.Line,
			GenerationMethod:     types.Search,
			SupportsAugmentation: true,
		}}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after attaching series to view (want/got): %s", diff)
		}

		err = store.RemoveSeriesFromView(ctx, series.SeriesID, view.ID)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		want = []types.InsightViewSeries{}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected result after removing series from view (want/got): %s", diff)
		}
		gotSeries, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: series.SeriesID, IncludeDeleted: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(gotSeries) == 0 || gotSeries[0].Enabled {
			t.Errorf("unexpected result: series does not exist or was not deleted after being removed from view")
		}
	})
}

func TestInsightStore_GetDataSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncate(time.Microsecond)
	groupByRepo := "repo"
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test empty", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected length of data series: %v", len(got))
		}
	})

	t.Run("test create and get series", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:             "unique-1",
			Query:                "query-1",
			OldestHistoricalAt:   now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			Enabled:              true,
			SampleIntervalUnit:   string(types.Month),
			GenerationMethod:     types.Search,
			GroupBy:              &groupByRepo,
			SupportsAugmentation: true,
		}
		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		want := []types.InsightSeries{created}

		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched insight data series want/got: %v", diff)
		}
	})

	t.Run("test create and get series just in time generation method", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:             "unique-1-gm-jit",
			Query:                "query-1-abc",
			OldestHistoricalAt:   now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			Enabled:              true,
			SampleIntervalUnit:   string(types.Month),
			JustInTime:           true,
			GenerationMethod:     types.Search,
			SupportsAugmentation: true,
		}
		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}
		want := []types.InsightSeries{created}

		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: "unique-1-gm-jit"})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched insight data series want/got: %v", diff)
		}
	})
}

func TestInsightStore_StampRecording(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test create and update stamp", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricalAt:  now.Add(-time.Hour * 24 * 365),
			LastRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LastSnapshotAt:      now,
			NextSnapshotAfter:   now,
			Enabled:             true,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		}
		created, err := store.CreateSeries(ctx, series)
		if err != nil {
			t.Fatal(err)
		}

		want := created
		want.LastRecordedAt = now
		want.NextRecordingAfter = time.Date(2020, 2, 5, 0, 0, 0, 0, time.UTC)

		got, err := store.StampRecording(ctx, created)
		if err != nil {
			return
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched updated recording stamp want/got: %v", diff)
		}
	})
}

func TestInsightStore_StampBackfillQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	series := types.InsightSeries{
		SeriesID:           "unique-1",
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
	created, err := store.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.StampBackfill(ctx, created)
	if err != nil {
		t.Fatal(err)
	}
	repoScope := "repo:scope"
	repoScopedSeries := types.InsightSeries{
		SeriesID:           "repoScoped",
		Query:              "query-2",
		OldestHistoricalAt: now.Add(-time.Hour * 24 * 365),
		LastRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LastSnapshotAt:     now,
		NextSnapshotAfter:  now,
		Enabled:            true,
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
		RepositoryCriteria: &repoScope,
	}
	repoScopedSeries, err = store.CreateSeries(ctx, repoScopedSeries)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.StampBackfill(ctx, repoScopedSeries)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("test only incomplete", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			BackfillNotQueued: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		want := 0
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched not queued backfill_stamp count want/got: %v", diff)
		}
	})
	t.Run("test get all", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		want := 2
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched get all count want/got: %v", diff)
		}
	})
	t.Run("test global only", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			GlobalOnly: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		wantCount := 1
		want := series.SeriesID
		if diff := cmp.Diff(wantCount, len(got)); diff != "" {
			t.Errorf("mismatched global only count want/got: %v", diff)
		}
		if diff := cmp.Diff(want, got[0].SeriesID); diff != "" {
			t.Errorf("mismatched global only seriesID want/got: %v", diff)
		}
	})
}

func TestInsightStore_StampBackfillCompleted(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	series := types.InsightSeries{
		SeriesID:           "unique-1",
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
	_, err := store.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SetSeriesBackfillComplete(ctx, "unique-1", now)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("test only incomplete", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{
			BackfillNotComplete: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		want := 0
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched updated backfill_stamp count want/got: %v", diff)
		}
	})
	t.Run("test get all", func(t *testing.T) {
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		want := 1
		if diff := cmp.Diff(want, len(got)); diff != "" {
			t.Errorf("mismatched updated backfill_stamp count want/got: %v", diff)
		}
	})
}

func TestSetSeriesEnabled(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("start enabled set disabled set enabled", func(t *testing.T) {
		created, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:           "series1",
			Query:              "quer1",
			CreatedAt:          now,
			OldestHistoricalAt: now,
			LastRecordedAt:     now,
			NextRecordingAfter: now,
			LastSnapshotAt:     now,
			NextSnapshotAfter:  now,
			BackfillQueuedAt:   now,
			SampleIntervalUnit: string(types.Month),
			GenerationMethod:   types.Search,
		})
		if err != nil {
			t.Fatal(err)
		}
		if !created.Enabled {
			t.Errorf("series is disabled")
		}
		// set the series from enabled -> disabled
		err = store.SetSeriesEnabled(ctx, created.SeriesID, false)
		if err != nil {
			t.Fatal(err)
		}
		got, err := store.GetDataSeries(ctx, GetDataSeriesArgs{IncludeDeleted: true, SeriesID: created.SeriesID})
		if err != nil {
			t.Fatal()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching data series")
		}
		if got[0].Enabled {
			t.Errorf("series is enabled but should be disabled")
		}

		// set the series from disabled -> enabled
		err = store.SetSeriesEnabled(ctx, created.SeriesID, true)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDataSeries(ctx, GetDataSeriesArgs{IncludeDeleted: true, SeriesID: created.SeriesID})
		if err != nil {
			t.Fatal()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching data series")
		}
		if !got[0].Enabled {
			t.Errorf("series is enabled but should be disabled")
		}
	})
}

func TestFindMatchingSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CreatedAt:           now,
		OldestHistoricalAt:  now,
		LastRecordedAt:      now,
		NextRecordingAfter:  now,
		LastSnapshotAt:      now,
		NextSnapshotAfter:   now,
		BackfillQueuedAt:    now,
		SampleIntervalUnit:  string(types.Week),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("find a matching series when one exists", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 1", StepIntervalUnit: string(types.Week), StepIntervalValue: 1})
		if err != nil {
			t.Fatal(err)
		}
		autogold.ExpectFile(t, gotSeries, autogold.ExportedOnly())
		autogold.Expect(true).Equal(t, gotFound)
	})
	t.Run("find no matching series when none exist", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 2", StepIntervalUnit: string(types.Week), StepIntervalValue: 1})
		if err != nil {
			t.Fatal(err)
		}
		autogold.ExpectFile(t, gotSeries, autogold.ExportedOnly())
		autogold.Expect(false).Equal(t, gotFound)
	})
	t.Run("match capture group series", func(t *testing.T) {
		_, err := store.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "series id capture group",
			Query:                      "query 1",
			CreatedAt:                  now,
			OldestHistoricalAt:         now,
			LastRecordedAt:             now,
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			BackfillQueuedAt:           now,
			SampleIntervalUnit:         string(types.Week),
			SampleIntervalValue:        1,
			GeneratedFromCaptureGroups: true,
			GenerationMethod:           types.SearchCompute,
		})
		if err != nil {
			t.Fatal(err)
		}
		gotSeries, gotFound, err := store.FindMatchingSeries(ctx, MatchSeriesArgs{Query: "query 1", StepIntervalUnit: string(types.Week), StepIntervalValue: 1, GenerateFromCaptureGroups: true})
		if err != nil {
			t.Fatal(err)
		}
		autogold.ExpectFile(t, gotSeries, autogold.ExportedOnly())
		autogold.Expect(true).Equal(t, gotFound)
	})
}

func TestUpdateFrontendSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncate(time.Microsecond)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CreatedAt:           now,
		OldestHistoricalAt:  now,
		LastRecordedAt:      now,
		NextRecordingAfter:  now,
		LastSnapshotAt:      now,
		NextSnapshotAfter:   now,
		BackfillQueuedAt:    now,
		SampleIntervalUnit:  string(types.Week),
		SampleIntervalValue: 1,
		GenerationMethod:    types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("updates a series", func(t *testing.T) {
		gotBeforeUpdate, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]types.InsightSeries{{
			ID:                   1,
			SeriesID:             "series id 1",
			Query:                "query 1",
			CreatedAt:            now,
			OldestHistoricalAt:   now,
			LastRecordedAt:       now,
			NextRecordingAfter:   now,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			Enabled:              true,
			SampleIntervalUnit:   "WEEK",
			SampleIntervalValue:  1,
			GenerationMethod:     "search",
			SupportsAugmentation: true,
		}}).Equal(t, gotBeforeUpdate)

		err = store.UpdateFrontendSeries(ctx, UpdateFrontendSeriesArgs{
			SeriesID:          "series id 1",
			Query:             "updated query!",
			StepIntervalUnit:  string(types.Month),
			StepIntervalValue: 5,
		})
		if err != nil {
			t.Fatal(err)
		}
		gotAfterUpdate, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]types.InsightSeries{{
			ID:                   1,
			SeriesID:             "series id 1",
			Query:                "updated query!",
			CreatedAt:            now,
			OldestHistoricalAt:   now,
			LastRecordedAt:       now,
			NextRecordingAfter:   now,
			LastSnapshotAt:       now,
			NextSnapshotAfter:    now,
			Enabled:              true,
			SampleIntervalUnit:   "MONTH",
			SampleIntervalValue:  5,
			GenerationMethod:     "search",
			SupportsAugmentation: true,
		}}).Equal(t, gotAfterUpdate)
	})
}

func TestGetReferenceCount(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'test title', 'test description', 'unique-1'),
									       (2, 'test title 2', 'test description 2', 'unique-2'),
										   (3, 'test title 3', 'test description 3', 'unique-3')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (id, title)
		VALUES (1, 'dashboard 1'), (2, 'dashboard 2'), (3, 'dashboard 3');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
									VALUES  (1, 1),
											(2, 1),
											(3, 1),
											(2, 2);`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("finds a single reference", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(referenceCount).Equal(t, 1)
	})
	t.Run("finds 3 references", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(referenceCount).Equal(t, 3)
	})
	t.Run("finds no references", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 3)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(referenceCount).Equal(t, 0)
	})
}

func TestGetSoftDeletedSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	deletedSeriesId := "soft_deleted"
	_, err := store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           deletedSeriesId,
		Query:              "deleteme",
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateSeries(ctx, types.InsightSeries{
		SeriesID:           "not_deleted",
		Query:              "keepme",
		SampleIntervalUnit: string(types.Month),
		GenerationMethod:   types.Search,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.SetSeriesEnabled(ctx, deletedSeriesId, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.GetSoftDeletedSeries(ctx, time.Now().AddDate(0, 0, 1)) // add some time just so the test can be ahead of the time the series was marked deleted
	if err != nil {
		t.Fatal(err)
	}
	autogold.Expect([]string{"soft_deleted"}).Equal(t, got)
}

func TestGetUnfrozenInsightCount(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := NewInsightStore(insightsDB)
	ctx := context.Background()

	t.Run("returns 0 if there are no insights", func(t *testing.T) {
		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(globalCount).Equal(t, 0)
		autogold.Expect(totalCount).Equal(t, 0)
	})
	t.Run("returns count for unfrozen insights not attached to dashboards", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'unattached insight', 'test description', 'unique-1', false)`)
		if err != nil {
			t.Fatal(err)
		}

		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(globalCount).Equal(t, 0)
		autogold.Expect(totalCount).Equal(t, 1)
	})
	t.Run("returns correct counts for unfrozen insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (2, 'private insight 2', 'test description', 'unique-2', true),
											   (3, 'org insight 1', 'test description', 'unique-3', false),
											   (4, 'global insight 1', 'test description', 'unique-4', false),
											   (5, 'global insight 2', 'test description', 'unique-5', false),
											   (6, 'global insight 3', 'test description', 'unique-6', true)`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (id, title)
										VALUES (1, 'private dashboard 1'),
											   (2, 'org dashboard 1'),
										 	   (3, 'global dashboard 1'),
										 	   (4, 'global dashboard 2');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
										VALUES  (1, 2),
												(2, 3),
												(3, 4),
												(4, 5),
												(4, 6);`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (id, dashboard_id, user_id, org_id, global)
										VALUES  (1, 1, 1, NULL, NULL),
												(2, 2, NULL, 1, NULL),
												(3, 3, NULL, NULL, TRUE),
												(4, 4, NULL, NULL, TRUE);`)
		if err != nil {
			t.Fatal(err)
		}

		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(globalCount).Equal(t, 2)
		autogold.Expect(totalCount).Equal(t, 4)
	})
}

func TestUnfreezeGlobalInsights(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := NewInsightStore(insightsDB)
	ctx := context.Background()

	t.Run("does nothing if there are no insights", func(t *testing.T) {
		err := store.UnfreezeGlobalInsights(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(globalCount).Equal(t, 0)
		autogold.Expect(totalCount).Equal(t, 0)
	})
	t.Run("does not unfreeze anything if there are no global insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'private insight 1', 'test description', 'unique-1', true),
											   (2, 'org insight 1', 'test description', 'unique-2', true),
											   (3, 'unattached insight', 'test description', 'unique-3', true);`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard (id, title)
										VALUES (1, 'private dashboard 1'),
											   (2, 'org dashboard 1'),
										 	   (3, 'global dashboard 1'),
										 	   (4, 'global dashboard 2');`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
										VALUES  (1, 1),
												(2, 2);`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (id, dashboard_id, user_id, org_id, global)
										VALUES  (1, 1, 1, NULL, NULL),
												(2, 2, NULL, 1, NULL),
												(3, 3, NULL, NULL, TRUE),
												(4, 4, NULL, NULL, TRUE);`)
		if err != nil {
			t.Fatal(err)
		}

		err = store.UnfreezeGlobalInsights(ctx, 3)
		if err != nil {
			t.Fatal(err)
		}
		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}

		autogold.Expect(globalCount).Equal(t, 0)
		autogold.Expect(totalCount).Equal(t, 0)
	})
	t.Run("unfreezes 2 global insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (4, 'global insight 1', 'test description', 'unique-4', true),
											   (5, 'global insight 2', 'test description', 'unique-5', true),
											   (6, 'global insight 3', 'test description', 'unique-6', true)`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
										VALUES  (3, 4),
												(3, 5),
												(4, 6);`)
		if err != nil {
			t.Fatal(err)
		}

		err = store.UnfreezeGlobalInsights(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		globalCount, totalCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect(globalCount).Equal(t, 2)
		autogold.Expect(totalCount).Equal(t, 2)
	})
}

func TestIncrementBackfillAttempts(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)

	_, err := insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
									VALUES (1, 'test title', 'test description', 'unique-1', false),
									       (2, 'test title 2', 'test description 2', 'unique-2', true)`)
	if err != nil {
		t.Fatal(err)
	}

	// assign some global grants just so the test can immediately fetch the created views
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_grants (insight_view_id, global)
									VALUES (1, true),
									       (2, true)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_series (series_id, query, created_at, oldest_historical_at, last_recorded_at,
                            next_recording_after, last_snapshot_at, next_snapshot_after, deleted_at, generation_method,backfill_attempts)
                            VALUES ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'search',0),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'search',1),
									('series-id-3', 'query-3', $1, $1, $1, $1, $1, $1, null, 'search',2);`, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
									VALUES (1, 1, 'label1', 'color1'),
											(1, 2, 'label2', 'color2'),
											(2, 2, 'second-label-2', 'second-color-2'),
											(2, 3, 'label3', 'color-2');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	store := NewInsightStore(insightsDB)

	all, err := store.GetDataSeries(ctx, GetDataSeriesArgs{})
	if err != nil {
		t.Fatal(err)
	}
	for _, series := range all {
		store.IncrementBackfillAttempts(context.Background(), series)
	}

	cases := []struct {
		seriesID string
		want     autogold.Value
	}{
		{"series-id-1", autogold.Expect(int32(1))},
		{"series-id-2", autogold.Expect(int32(2))},
		{"series-id-3", autogold.Expect(int32(3))},
	}

	for _, tc := range cases {
		t.Run(tc.seriesID, func(t *testing.T) {
			series, err := store.GetDataSeries(ctx, GetDataSeriesArgs{SeriesID: tc.seriesID})
			if err != nil {
				t.Fatal(err)
			}

			got := series[0].BackfillAttempts
			tc.want.Equal(t, got)
		})
	}
}

func TestHardDeleteSeries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsdb := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	postgres := database.NewDB(logger, dbtest.NewDB(t))
	permStore := NewInsightPermissionStore(postgres)
	insightStore := NewInsightStore(insightsdb)
	timeseriesStore := NewWithClock(insightsdb, permStore, clock)

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
	got, err := insightStore.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 {
		t.Errorf("expected first series to have id 1")
	}
	series.SeriesID = "series2" // copy to make a new one
	got, err = insightStore.CreateSeries(ctx, series)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 2 {
		t.Errorf("expected second series to have id 2")
	}

	err = timeseriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{
		{
			InsightSeriesID: 1,
			RecordingTimes:  []types.RecordingTime{{Timestamp: now}},
		},
		{
			InsightSeriesID: 2,
			RecordingTimes:  []types.RecordingTime{{Timestamp: now}},
		},
	})
	if err != nil {
		t.Error(err)
	}

	if err = insightStore.HardDeleteSeries(ctx, "series1"); err != nil {
		t.Fatal(err)
	}

	getInsightSeries := func(ctx context.Context, timeseriesStore *Store, seriesId string) bool {
		q := sqlf.Sprintf("select count(*) from insight_series where series_id = %s;", seriesId)
		val, err := basestore.ScanInt(timeseriesStore.QueryRow(ctx, q))
		if err != nil {
			t.Fatal(err)
		}
		return val == 1
	}

	getTimesCountforSeries := func(ctx context.Context, timeseriesStore *Store, seriesId int) int {
		q := sqlf.Sprintf("select count(*) from insight_series_recording_times where insight_series_id = %s;", seriesId)
		val, err := basestore.ScanInt(timeseriesStore.QueryRow(ctx, q))
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	if getInsightSeries(ctx, timeseriesStore, "series1") {
		t.Errorf("expected series1 to be deleted")
	}
	if getTimesCountforSeries(ctx, timeseriesStore, 1) != 0 {
		t.Errorf("expected 0 recording times to remain for series1")
	}

	if !getInsightSeries(ctx, timeseriesStore, "series2") {
		t.Errorf("expected series2 to be there")
	}
	if getTimesCountforSeries(ctx, timeseriesStore, 2) != 1 {
		t.Errorf("expected 1 recording times to remain for series2")
	}
}
