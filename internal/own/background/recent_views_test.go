package background

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRecentViewsIndexer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Assertion function.
	assertSummaries := func(summariesCount, expectedCount int) {
		t.Helper()
		opts := database.ListRecentViewSignalOpts{IncludeAllPaths: true}
		summaries, err := db.RecentViewSignal().List(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, summaries, summariesCount)
	}

	// Creating a worker.
	indexer := newRecentViewsIndexer(db, logger)

	// Mock authz checker
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.EnabledForRepoFunc.SetDefaultReturn(false, nil)

	// Dry run of handling: we should not have any summaries yet.
	err = indexer.handle(ctx, checker)
	require.NoError(t, err)
	// Assertions are in the loop over listed summaries -- it won't error out when
	// there are 0 summaries.
	assertSummaries(0, -1)

	// Adding events.
	insertEvents(ctx, t, db)

	sortSummaries := func(ss []database.RecentViewSummary) {
		sort.Slice(ss, func(i, j int) bool { return ss[i].FilePathID < ss[j].FilePathID })
	}

	// expectedSummaries is the recent view summaries we intend to see
	// for two relevant events inserted by given number of runs of `insertEvents`.
	expectedSummaries := func(handlerRuns int) []database.RecentViewSummary {
		rs, err := db.QueryContext(ctx, "SELECT id, absolute_path FROM repo_paths")
		require.NoError(t, err)
		defer rs.Close()
		// `insertEvents` inserts two relevant events for paths:
		// - cmd/gitserver/internal/main.go
		// - cmd/gitserver/internal/patch.go
		// Since these are in the same directory:
		// - every summary record that indicates a file, will have a count
		//   corresponding to the number of handlerRuns
		// - and every record corresponding to a parent/ancestor directory
		//   will have a count twice as big.
		// We recognize whether a path is a leaf file, by checking for .go suffix.
		var summaries []database.RecentViewSummary
		for rs.Next() {
			var id int
			var absolutePath string
			require.NoError(t, rs.Scan(&id, &absolutePath))
			count := handlerRuns
			if !strings.HasSuffix(absolutePath, ".go") {
				count = 2 * count
			}
			summaries = append(summaries, database.RecentViewSummary{
				UserID:     1,
				FilePathID: id,
				ViewsCount: count,
			})
		}
		return summaries
	}

	// First round of handling: we should have all counts equal to 1.
	err = indexer.handle(ctx, checker)
	require.NoError(t, err)
	got, err := db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{IncludeAllPaths: true})
	require.NoError(t, err)
	want := expectedSummaries(1)
	sortSummaries(got)
	sortSummaries(want)
	assert.Equal(t, want, got)

	// Now we can insert some more events.
	insertEvents(ctx, t, db)

	// Second round of handling: we should have all counts equal to 2.
	err = indexer.handle(ctx, checker)
	require.NoError(t, err)
	got, err = db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{IncludeAllPaths: true})
	require.NoError(t, err)
	want = expectedSummaries(2)
	sortSummaries(got)
	sortSummaries(want)
	assert.Equal(t, want, got)

	// Now we can insert some more events, but the checker will now caregorize this repo as having subrepo perms enabled
	insertEvents(ctx, t, db)
	checker.EnabledForRepoFunc.SetDefaultReturn(true, nil)

	// Third round of handling: we should have all counts equal to 2.
	err = indexer.handle(ctx, checker)
	require.NoError(t, err)
	got, err = db.RecentViewSignal().List(ctx, database.ListRecentViewSignalOpts{IncludeAllPaths: true})
	require.NoError(t, err)
	// we expect the summary to be no different than before since all new view events should be captured
	// due to the repo having sub-repo permissions enabled.
	want = expectedSummaries(2)
	sortSummaries(got)
	sortSummaries(want)
	assert.Equal(t, want, got)

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
