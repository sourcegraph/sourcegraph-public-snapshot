package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
		clock := func() time.Time { return now }
		insightsDB := dbtest.NewInsightsDB(t)
		postgres := dbtest.NewDB(t)
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
