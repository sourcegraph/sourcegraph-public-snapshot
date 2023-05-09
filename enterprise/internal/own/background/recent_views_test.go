package background

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecentViewsIndexer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Enabling a feature flag.
	_, err = db.FeatureFlags().CreateBool(ctx, "own-background-index-repo-recent-views", true)
	require.NoError(t, err)

	// Assertion function.
	assertSummaries := func(summariesCount, expectedCount int) {
		t.Helper()
		summaries, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{})
		require.NoError(t, err)
		assert.Len(t, summaries, summariesCount)
		for _, summary := range summaries {
			assert.Equal(t, expectedCount, summary.ViewsCount)
		}
	}

	// Creating a worker.
	indexer := newRecentViewsIndexer(db, logger)

	// Dry run of handling: we should not have any summaries yet.
	err = indexer.Handle(ctx)
	require.NoError(t, err)
	// Assertions are in the loop over listed summaries -- it won't error out when
	// there are 0 summaries.
	assertSummaries(0, -1)

	// Adding events.
	insertEvents(ctx, t, db)

	// First round of handling: we should have all counts equal to 1.
	err = indexer.Handle(ctx)
	require.NoError(t, err)
	assertSummaries(2, 1)

	// Now we can insert some more events.
	insertEvents(ctx, t, db)

	// Second round of handling: we should have all counts equal to 2.
	err = indexer.Handle(ctx)
	require.NoError(t, err)
	assertSummaries(2, 2)
}

func insertEvents(ctx context.Context, t *testing.T, db database.DB) {
	t.Helper()
	events := []*database.Event{
		{
			UserID: 1,
			Name:   "SearchResultsQueried",
			URL:    "http://sourcegraph.com",
			Source: "test",
		}, {
			UserID: 1,
			Name:   "codeintel",
			URL:    "http://sourcegraph.com",
			Source: "test",
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			URL:            "http://sourcegraph.com",
			Source:         "test",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID: 1,
			Name:   "SearchResultsQueried",
			URL:    "http://sourcegraph.com",
			Source: "test",
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			URL:            "http://sourcegraph.com",
			Source:         "test",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/main.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
	}

	require.NoError(t, db.EventLogs().BulkInsert(ctx, events))
}
