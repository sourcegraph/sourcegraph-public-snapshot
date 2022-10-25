package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

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
		logger := logtest.Scoped(t)
		clock := func() time.Time { return now }
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
		postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
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
