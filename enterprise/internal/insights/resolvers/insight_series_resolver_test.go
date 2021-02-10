package resolvers

import (
	"context"
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
}
