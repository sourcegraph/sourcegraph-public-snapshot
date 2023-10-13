package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestResolver_Insights just checks that root resolver setup and getting an insights connection
// does not result in any errors. It is a pretty minimal test.
func TestResolver_Insights(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now }
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	postgres := database.NewDB(logger, dbtest.NewDB(t))
	resolver := newWithClock(insightsDB, postgres, clock)

	insightsConnection, err := resolver.InsightViews(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if insightsConnection == nil {
		t.Fatal("got nil")
	}
}
