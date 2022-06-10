package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
		clock := func() time.Time { return now }
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(t))
		postgres := database.NewDB(dbtest.NewDB(t))
		resolver := newWithClock(insightsDB, postgres, clock)

		insightMetadataStore := store.NewMockInsightMetadataStore()
		insightMetadataStore.GetMappedFunc.SetDefaultReturn([]types.Insight{
			{
				UniqueID:    "unique1",
				Title:       "title1",
				Description: "desc1",
				Series: []types.InsightViewSeries{
					{
						UniqueID:           "unique1",
						SeriesID:           "1234567",
						Title:              "title1",
						Description:        "desc1",
						Query:              "query1",
						CreatedAt:          now,
						OldestHistoricalAt: now,
						LastRecordedAt:     now,
						NextRecordingAfter: now,
						Label:              "label1",
						LineColor:          "color1",
					},
				},
			},
		}, nil)
		resolver.insightMetadataStore = insightMetadataStore

		// Create the insights connection resolver and query series.
		conn, err := resolver.Insights(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		nodes, err := conn.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var series [][]graphqlbackend.InsightSeriesResolver
		for _, node := range nodes {
			series = append(series, node.Series())
		}
		return ctx, series
	}

	t.Run("Points", func(t *testing.T) {
		ctx, insights := testSetup(t)
		autogold.Want("insights length", int(1)).Equal(t, len(insights))

		autogold.Want("insights[0].length", int(1)).Equal(t, len(insights[0]))

		// Issue a query against the actual DB.
		points, err := insights[0][0].Points(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("insights[0][0].Points", []graphqlbackend.InsightsDataPointResolver{}).Equal(t, points)

	})
}

func TestGetSortedCaptureGroups(t *testing.T) {
	seriesId := "series_id"
	getPoint := func(month time.Month, value float64, capture string) store.SeriesPoint {
		return store.SeriesPoint{
			SeriesID: seriesId,
			Time:     time.Date(2021, month, 5, 1, 1, 1, 1, time.UTC),
			Value:    value,
			Metadata: []byte{},
			Capture:  &capture,
		}
	}
	captureGroups := map[string][]store.SeriesPoint{
		"v0.2.1": {getPoint(time.July, 50, "v0.2.1")},
		"1.2.3":  {getPoint(time.March, 70, "1.2.3")},
		"earliest date added": {
			getPoint(time.January, 50, "earliest date added"),
			getPoint(time.February, 50, "earliest date added"),
			getPoint(time.March, 50, "earliest date added"),
			getPoint(time.April, 50, "earliest date added"),
			getPoint(time.May, 50, "earliest date added"),
		},
		"latest date added": {
			getPoint(time.December, 50, "latest date added"),
		},
		"least results": {
			getPoint(time.July, 500, "least results"),
			getPoint(time.August, 100, "least results"),
			getPoint(time.September, 10, "least results"),
		},
		"most results": {
			getPoint(time.July, 10, "most results"),
			getPoint(time.August, 100, "most results"),
			getPoint(time.September, 500, "most results"),
		},
	}

	getOptions := func(sortOptions *types.SeriesSortOptions, limit *int32) types.SeriesDisplayOptions {
		return types.SeriesDisplayOptions{
			SortOptions: sortOptions,
			Limit:       limit,
		}
	}
	nilOptions := getOptions(nil, nil)

	getDefinition := func(mode types.SeriesSortMode, direction types.SeriesSortDirection, limit int32) types.InsightViewSeries {
		return types.InsightViewSeries{
			SeriesSortMode:      &mode,
			SeriesSortDirection: &direction,
			SeriesLimit:         &limit,
		}
	}

	t.Run("sorts by asc lexicographical", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.Lexicographical, types.Asc, 6), captureGroups)
		if diff := cmp.Diff([]string{"v0.2.1", "1.2.3", "earliest date added", "latest date added", "least results", "most results"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("sorts by desc lexicographical", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.Lexicographical, types.Desc, 6), captureGroups)
		if diff := cmp.Diff([]string{"most results", "least results", "latest date added", "earliest date added", "1.2.3", "v0.2.1"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("sorts by asc date added", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.DateAdded, types.Asc, 6), captureGroups)
		if diff := cmp.Diff([]string{"earliest date added", "1.2.3", "v0.2.1", "least results", "most results", "latest date added"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("sorts by desc date added", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.DateAdded, types.Desc, 6), captureGroups)
		if diff := cmp.Diff([]string{"latest date added", "v0.2.1", "least results", "most results", "1.2.3", "earliest date added"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("sorts by asc result count", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.ResultCount, types.Asc, 6), captureGroups)
		if diff := cmp.Diff([]string{"least results", "v0.2.1", "earliest date added", "latest date added", "1.2.3", "most results"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("sorts by desc result count", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, getDefinition(types.ResultCount, types.Desc, 5), captureGroups)
		if diff := cmp.Diff([]string{"most results", "1.2.3", "v0.2.1", "earliest date added", "latest date added", "least results"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(5), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("uses override options over definition", func(t *testing.T) {
		optionsLimit := int32(3)
		sorted, limit := getSortedCaptureGroups(getOptions(&types.SeriesSortOptions{Mode: types.Lexicographical, Direction: types.Desc}, &optionsLimit), getDefinition(types.ResultCount, types.Desc, 6), captureGroups)
		if diff := cmp.Diff([]string{"most results", "least results", "latest date added", "earliest date added", "1.2.3", "v0.2.1"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(3), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
	t.Run("uses defaults when no options or definitions are set", func(t *testing.T) {
		sorted, limit := getSortedCaptureGroups(nilOptions, types.InsightViewSeries{}, captureGroups)
		// defaults to desc result count
		if diff := cmp.Diff([]string{"most results", "1.2.3", "v0.2.1", "earliest date added", "latest date added", "least results"}, sorted); diff != "" {
			t.Errorf("unexpected sort order (want/got): %v", diff)
		}
		if diff := cmp.Diff(int32(6), limit); diff != "" {
			t.Errorf("unexpected limit (want/got): %v", diff)
		}
	})
}
