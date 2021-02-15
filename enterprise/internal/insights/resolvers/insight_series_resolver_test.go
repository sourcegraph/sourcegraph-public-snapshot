package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver, *store.MockInterface, func()) {
		// Setup the GraphQL resolver.
		ctx := backend.WithAuthzBypass(context.Background())
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now }
		timescale, cleanup := insightsdbtesting.TimescaleDB(t)
		postgres := dbtesting.GetDB(t)
		resolver := newWithClock(timescale, postgres, clock)

		// Create a mock store, delegating any un-mocked methods to the DB store.
		dbStore := resolver.store
		mockStore := store.NewMockInterfaceFrom(dbStore)
		resolver.store = mockStore

		// Create the insights connection resolver and query series.
		conn, err := resolver.Insights(ctx)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}
		conn.(*insightConnectionResolver).mocksSettingsGetLatest = func(ctx context.Context, subject api.SettingsSubject) (*api.Settings, error) {
			if !subject.Site { // TODO: future: site is an extremely poor name for "global settings", we should change this.
				t.Fatal("expected only to request settings from global user settings")
			}
			return testRealGlobalSettings, nil
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

	t.Run("metadata", func(t *testing.T) {
		_, insights, _, cleanup := testSetup(t)
		defer cleanup()
		autogold.Want("insights length", int(2)).Equal(t, len(insights))

		autogold.Want("insights[0].length", int(2)).Equal(t, len(insights[0]))
		autogold.Want("insights[0].series[0].Label", "fmt.Errorf").Equal(t, insights[0][0].Label())
		autogold.Want("insights[0].series[1].Label", "printf").Equal(t, insights[0][1].Label())

		autogold.Want("insights[1].length", int(2)).Equal(t, len(insights[1]))
		autogold.Want("insights[1].series[0].Label", "exec").Equal(t, insights[1][0].Label())
		autogold.Want("insights[1].series[1].Label", "close").Equal(t, insights[1][1].Label())
	})

	t.Run("Points", func(t *testing.T) {
		ctx, insights, mock, cleanup := testSetup(t)
		defer cleanup()
		autogold.Want("insights length", int(2)).Equal(t, len(insights))

		autogold.Want("insights[0].length", int(2)).Equal(t, len(insights[0]))
		autogold.Want("insights[1].length", int(2)).Equal(t, len(insights[1]))

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
			autogold.Want("insights[0][0].Points store opts", "{SeriesID:<nil> From:2006-01-02 15:04:05 +0000 UTC To:2006-01-03 15:04:05 +0000 UTC Limit:0}").Equal(t, fmt.Sprintf("%+v", opts))
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
		autogold.Want("insights[0][0].Points mocked", "[{p:{Time:{wall:0 ext:63271811045 loc:<nil>} Value:1 Metadata:[]}} {p:{Time:{wall:0 ext:63271811045 loc:<nil>} Value:2 Metadata:[]}} {p:{Time:{wall:0 ext:63271811045 loc:<nil>} Value:3 Metadata:[]}}]").Equal(t, fmt.Sprintf("%+v", points))
	})
}
