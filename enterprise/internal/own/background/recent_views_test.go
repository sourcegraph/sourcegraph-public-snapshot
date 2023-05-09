package background

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

	// Creating a worker.
	observationCtx := &observation.TestContext
	mockIndexInterval = time.Second * 3
	cleanMock := func() { mockIndexInterval = 0 }
	t.Cleanup(cleanMock)
	worker := NewOwnRecentViewsIndexer(db, observationCtx)
	go worker.Start()
	t.Cleanup(worker.Stop)

	// Adding events.
	insertEvents(ctx, t, db)

	// 30 seconds timeout in case of something being wrong, to terminate the test
	// run.
	timeout := time.After(30 * time.Second)
	// First round is to wait for 2 summaries to be inserted to the database.
	//
	// Second round is to wait for the same summaries to be updated (their count
	// should be incremented).
	round := 1

	assertSummaries := func() (needToWait bool) {
		summaries, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{})
		require.NoError(t, err)
		if round == 1 && len(summaries) == 2 {
			// We don't need to check all summary fields because it is covered in
			// `recent_view_signal_test.go`.
			for _, summary := range summaries {
				// Count of views is equal to round number.
				assert.Equal(t, round, summary.ViewsCount)
			}
			// We are fine to add more events and go to next round.
			insertEvents(ctx, t, db)
			round++
		} else if round == 2 {
			// In round 2, we're waiting for the second iteration of worker which should
			// increment the counts.
			sum := 0
			for _, summary := range summaries {
				// Count of views is equal to round number.
				sum += summary.ViewsCount
			}
			// If both counts have been incremented -- we go to next round, which leads to
			// successful end of this test.
			if sum == 4 {
				round++
				return false
			}
		}
		return true
	}
loop:
	for {
		// We need to do assertions for rounds 1 and 2. Round 3 is success -- exit the
		// loop.
		switch round {
		case 1, 2:
			needToWait := assertSummaries()
			if needToWait {
				// Waiting for the summaries to be inserted.
				time.Sleep(1 * time.Second)
			}
		default:
			break loop
		}

		// Checking for timeout.
		select {
		case <-timeout:
			t.Fatal("Event logs are not processing or processing takes too much time.")
		default:
			continue loop
		}
	}
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
