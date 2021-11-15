package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver, *store.MockInterface, func()) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
		clock := func() time.Time { return now }
		timescale, cleanup := insightsdbtesting.TimescaleDB(t)
		postgres := dbtesting.GetDB(t)
		resolver := newWithClock(timescale, postgres, clock)

		// Create a mock store, delegating any un-mocked methods to the DB store.
		dbStore := resolver.timeSeriesStore
		mockStore := store.NewMockInterfaceFrom(dbStore)
		resolver.timeSeriesStore = mockStore

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
			cleanup()
			t.Fatal(err)
		}

		nodes, err := conn.Nodes(ctx)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}
		var series [][]graphqlbackend.InsightSeriesResolver
		for _, node := range nodes {
			series = append(series, node.Series())
		}
		return ctx, series, mockStore, cleanup
	}

	t.Run("Points", func(t *testing.T) {
		ctx, insights, mock, cleanup := testSetup(t)
		defer cleanup()
		autogold.Want("insights length", int(1)).Equal(t, len(insights))

		autogold.Want("insights[0].length", int(1)).Equal(t, len(insights[0]))

		args := &graphqlbackend.InsightsPointsArgs{
			From: &graphqlbackend.DateTime{Time: time.Now().Add(-7 * 24 * time.Hour)},
			To:   &graphqlbackend.DateTime{Time: time.Now()},
		}

		// Issue a query against the actual DB.
		points, err := insights[0][0].Points(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("insights[0][0].Points", []graphqlbackend.InsightsDataPointResolver{}).Equal(t, points)

		// Mock the store and confirm args got passed through as expected.
		args.From.Time, _ = time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
		args.To.Time, _ = time.Parse(time.RFC3339, "2006-01-03T15:04:05Z")
		mock.SeriesPointsFunc.SetDefaultHook(func(ctx context.Context, opts store.SeriesPointsOpts) ([]store.SeriesPoint, error) {
			json, err := json.Marshal(opts)
			if err != nil {
				t.Fatal(err)
			}
			autogold.Want("insights[0][0].Points store opts", `{"SeriesID":"1234567","RepoID":null,"Excluded":null,"Included":null,"IncludeRepoRegex":"","ExcludeRepoRegex":"","From":"2006-01-02T15:04:05Z","To":"2006-01-03T15:04:05Z","Limit":0}`).Equal(t, string(json))
			return []store.SeriesPoint{
				{Time: args.From.Time, Value: 1},
				{Time: args.From.Time, Value: 2},
				{Time: args.From.Time, Value: 3},
			}, nil
		})
		points, err = insights[0][0].Points(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("insights[0][0].Points mocked", "[{p:{SeriesID: Time:{wall:0 ext:63271811045 loc:<nil>} Value:1 Metadata:[]}} {p:{SeriesID: Time:{wall:0 ext:63271811045 loc:<nil>} Value:2 Metadata:[]}} {p:{SeriesID: Time:{wall:0 ext:63271811045 loc:<nil>} Value:3 Metadata:[]}}]").Equal(t, fmt.Sprintf("%+v", points))
	})
}
