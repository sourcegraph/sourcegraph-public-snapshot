package database

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRecentViewSignalStore_BuildAggregateFromEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating 2 users.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	_, err = db.Users().Create(ctx, NewUser{Username: "user2"})
	require.NoError(t, err)

	// Creating 2 repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"}, &types.Repo{ID: 2, Name: "github.com/sourcegraph/sourcegraph2"})
	require.NoError(t, err)

	// Creating ViewBlob events.
	events := []*Event{
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "enterprise/cmd/frontend/main.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "enterprise/cmd/frontend/main.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/sourcegraph2"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/sourcegraph2"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph2"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/not/found"}`),
		},
	}

	// Building signal aggregates.
	store := RecentViewSignalStoreWith(db, logger)
	err = store.BuildAggregateFromEvents(ctx, events)
	require.NoError(t, err)

	resolvePathsForRepo := func(ctx context.Context, db DB, repoID int) map[string]int {
		t.Helper()
		rows, err := db.QueryContext(ctx, "SELECT id, absolute_path FROM repo_paths WHERE repo_id = $1 AND absolute_path LIKE $2", repoID, "%.go")
		require.NoError(t, err)
		pathToID := make(map[string]int)
		for rows.Next() {
			var id int
			var path string
			err := rows.Scan(&id, &path)
			require.NoError(t, err)
			pathToID[path] = id
		}
		return pathToID
	}

	// Getting actual mapping of path to its ID for both repos.
	repo1PathToID := resolvePathsForRepo(ctx, db, 1)
	repo2PathToID := resolvePathsForRepo(ctx, db, 2)

	// Getting all RecentViewSummary entries from the DB and checking their
	// correctness.
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)

	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["cmd/gitserver/server/lock.go"], ViewsCount: 2})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["cmd/gitserver/server/patch.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["enterprise/cmd/frontend/main.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 2, FilePathID: repo1PathToID["enterprise/cmd/frontend/main.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 2, FilePathID: repo1PathToID["cmd/gitserver/server/lock.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 2, FilePathID: repo2PathToID["cmd/gitserver/server/patch.go"], ViewsCount: 2})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 2, FilePathID: repo2PathToID["cmd/gitserver/server/lock.go"], ViewsCount: 1})
}

func TestRecentViewSignalStore_BuildAggregateFromEvents_WithExcludedRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating 2 users.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	_, err = db.Users().Create(ctx, NewUser{Username: "user2"})
	require.NoError(t, err)

	// Creating 3 repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"}, &types.Repo{ID: 2, Name: "github.com/sourcegraph/pattern-repo-1337"}, &types.Repo{ID: 3, Name: "github.com/sourcegraph/pattern-repo-421337"})
	require.NoError(t, err)

	// Creating ViewBlob events.
	events := []*Event{
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         1,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "enterprise/cmd/frontend/main.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "enterprise/cmd/frontend/main.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/sourcegraph"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/pattern-repo-1337"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/patch.go", "repoName": "github.com/sourcegraph/pattern-repo-421337"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/sourcegraph/pattern-repo-421337"}`),
		},
		{
			UserID:         2,
			Name:           "ViewBlob",
			PublicArgument: json.RawMessage(`{"filePath": "cmd/gitserver/server/lock.go", "repoName": "github.com/not/found"}`),
		},
	}

	// Adding a config with excluded repos.
	configStore := SignalConfigurationStoreWith(db)
	err = configStore.UpdateConfiguration(ctx, UpdateSignalConfigurationArgs{Name: "recent-views", Enabled: true, ExcludedRepoPatterns: []string{"github.com/sourcegraph/pattern-repo%"}})
	require.NoError(t, err)
	err = configStore.UpdateConfiguration(ctx, UpdateSignalConfigurationArgs{Name: "recent-contributors", Enabled: true, ExcludedRepoPatterns: []string{"github.com/sourcegraph/sourcegraph"}})
	require.NoError(t, err)

	// Building signal aggregates.
	store := RecentViewSignalStoreWith(db, logger)
	err = store.BuildAggregateFromEvents(ctx, events)
	require.NoError(t, err)

	resolvePathsForRepo := func(ctx context.Context, db DB, repoID int) map[string]int {
		t.Helper()
		rows, err := db.QueryContext(ctx, "SELECT id, absolute_path FROM repo_paths WHERE repo_id = $1 AND absolute_path LIKE $2", repoID, "%.go")
		require.NoError(t, err)
		pathToID := make(map[string]int)
		for rows.Next() {
			var id int
			var path string
			err := rows.Scan(&id, &path)
			require.NoError(t, err)
			pathToID[path] = id
		}
		return pathToID
	}

	// Getting actual mapping of path to its ID.
	repo1PathToID := resolvePathsForRepo(ctx, db, 1)

	// Getting all RecentViewSummary entries from the DB and checking their
	// correctness.
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)

	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["cmd/gitserver/server/lock.go"], ViewsCount: 2})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["cmd/gitserver/server/patch.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 1, FilePathID: repo1PathToID["enterprise/cmd/frontend/main.go"], ViewsCount: 1})
	assert.Contains(t, summaries, RecentViewSummary{UserID: 2, FilePathID: repo1PathToID["enterprise/cmd/frontend/main.go"], ViewsCount: 1})

	// We shouldn't have any paths inserted for repos
	// "github.com/sourcegraph/pattern-repo-1337" and
	// "github.com/sourcegraph/pattern-repo-421337" because they are excluded.
	count, _, err := basestore.ScanFirstInt(db.QueryContext(context.Background(), "SELECT COUNT(*) FROM repo_paths WHERE repo_id IN (2, 3)"))
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestRecentViewSignalStore_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating a couple of paths.
	_, err = db.QueryContext(ctx, "INSERT INTO repo_paths (repo_id, absolute_path, parent_id) VALUES (1, '', NULL), (1, 'src', 1), (1, 'src/abc', 2)")
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	clearTable := func(ctx context.Context, db DB) {
		_, err = db.QueryContext(ctx, "DELETE FROM own_aggregate_recent_view")
		require.NoError(t, err)
	}

	t.Run("inserting initial signal", func(t *testing.T) {
		err = store.Insert(ctx, 1, 2, 10)
		require.NoError(t, err)
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		assert.Equal(t, 2, summaries[0].FilePathID)
		assert.Equal(t, 10, summaries[0].ViewsCount)
		clearTable(ctx, db)
	})

	t.Run("inserting multiple signals", func(t *testing.T) {
		err = store.Insert(ctx, 1, 2, 10)
		err = store.Insert(ctx, 1, 3, 20)
		require.NoError(t, err)
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
		require.NoError(t, err)
		assert.Len(t, summaries, 2)
		assert.Equal(t, 2, summaries[0].FilePathID)
		assert.Equal(t, 10, summaries[0].ViewsCount)
		assert.Equal(t, 3, summaries[1].FilePathID)
		assert.Equal(t, 20, summaries[1].ViewsCount)
		clearTable(ctx, db)
	})

	t.Run("inserting conflicting entry will update it", func(t *testing.T) {
		err = store.Insert(ctx, 1, 2, 10)
		require.NoError(t, err)
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		assert.Equal(t, 2, summaries[0].FilePathID)
		assert.Equal(t, 10, summaries[0].ViewsCount)

		// Inserting a conflicting entry.
		err = store.Insert(ctx, 1, 2, 100)
		require.NoError(t, err)
		summaries, err = store.List(ctx, ListRecentViewSignalOpts{})
		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		assert.Equal(t, 2, summaries[0].FilePathID)
		assert.Equal(t, 110, summaries[0].ViewsCount)
		clearTable(ctx, db)
	})
}

func storeFrom(t *testing.T, d DB) *basestore.Store {
	t.Helper()
	casted, ok := d.(*db)
	if !ok {
		t.Fatal("cannot cast DB down to retrieve store")
	}
	return casted.Store
}

func TestRecentViewSignalStore_InsertPaths(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating 4 paths.
	pathIDs, err := ensureRepoPaths(ctx, storeFrom(t, db), []string{
		"foo",
		"src/cde",
		// To also get parent and root ID.
		"src",
		"",
	}, 1)
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	err = store.InsertPaths(ctx, 1, map[int]int{
		pathIDs[0]: 100,  // file foo
		pathIDs[1]: 1000, // file src/cde
	})
	require.NoError(t, err)
	got, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)
	want := []RecentViewSummary{
		{
			UserID:     1,
			FilePathID: pathIDs[0], // foo
			ViewsCount: 100,        // Leaf: Return the views inserted for foo
		},
		{
			UserID:     1,
			FilePathID: pathIDs[1], // src/cde
			ViewsCount: 1000,       // Leaf: Return the views inserted for src/cde
		},
		{
			UserID:     1,
			FilePathID: pathIDs[2], // src
			ViewsCount: 1000,       // Sum for the only file with views - src/cde
		},
		{
			UserID:     1,
			FilePathID: pathIDs[3], // "" - root
			ViewsCount: 1000 + 100, // Sum for foo and src/cde
		},
	}
	sort.Slice(got, func(i, j int) bool { return got[i].FilePathID < got[j].FilePathID })
	sort.Slice(want, func(i, j int) bool { return want[i].FilePathID < want[j].FilePathID })
	assert.Equal(t, want, got)
}

func TestRecentViewSignalStore_InsertPaths_OverBatchSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating 5500 paths.
	var paths []string
	for i := 1; i <= 5500; i++ {
		paths = append(paths, fmt.Sprintf("src/file%d", i))
	}
	pathIDs, err := ensureRepoPaths(ctx, storeFrom(t, db), paths, 1)
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	counts := map[int]int{}
	for _, id := range pathIDs {
		counts[id] = 10
	}

	err = store.InsertPaths(ctx, 1, counts)
	require.NoError(t, err)
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)
	require.Len(t, summaries, 5502) // Two extra entries - repo root and 'src' directory
}

func TestRecentViewSignalStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating some paths.
	_, err = db.QueryContext(ctx, "INSERT INTO repo_paths (repo_id, absolute_path, parent_id) VALUES (1, '', NULL), (1, 'src', 1), (1, 'src/abc', 2), (1, 'src/cde', 2)")
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	// We need a determined order of inserted signals, hence using a singular insert.
	for _, path := range []int{1, 2, 3, 4} {
		// count = 10^path
		require.NoError(t, store.Insert(ctx, 1, path, int(math.Pow(float64(10), float64(path)))))
	}

	// As IDs of signals aren't returned, we can rely on counts because of strict
	// mapping.
	allCounts := []int{10, 100, 1000, 10000}
	testCases := map[string]struct {
		opts              ListRecentViewSignalOpts
		expectedCounts    []int
		expectedNoEntries bool
	}{
		"listing everything without opts": {
			opts:           ListRecentViewSignalOpts{},
			expectedCounts: allCounts,
		},
		"filter by viewer ID": {
			opts:           ListRecentViewSignalOpts{ViewerUserID: 1},
			expectedCounts: allCounts,
		},
		"filter by viewer ID which isn't present": {
			opts:              ListRecentViewSignalOpts{ViewerUserID: 2},
			expectedNoEntries: true,
		},
		"filter by repo ID": {
			opts:           ListRecentViewSignalOpts{RepoID: 1},
			expectedCounts: allCounts,
		},
		"filter by repo ID which isn't present": {
			opts:              ListRecentViewSignalOpts{RepoID: 2},
			expectedNoEntries: true,
		},
		"filter by path": {
			opts:           ListRecentViewSignalOpts{Path: "src/cde"},
			expectedCounts: allCounts[len(allCounts)-1:],
		},
		"filter by path which isn't present": {
			opts:              ListRecentViewSignalOpts{Path: "lol"},
			expectedNoEntries: true,
		},
		"limit, offset": {
			opts:           ListRecentViewSignalOpts{LimitOffset: &LimitOffset{Limit: 2, Offset: 1}},
			expectedCounts: allCounts[1:4],
		},
		"all options": {
			opts:           ListRecentViewSignalOpts{ViewerUserID: 1, RepoID: 1, Path: "src", LimitOffset: &LimitOffset{Limit: 1}},
			expectedCounts: allCounts[1:2],
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			gotSummaries, err := store.List(ctx, test.opts)
			require.NoError(t, err)
			if test.expectedNoEntries {
				assert.Empty(t, gotSummaries)
				return
			}
			for idx, summary := range gotSummaries {
				assert.Equal(t, test.expectedCounts[idx], summary.ViewsCount)
			}
		})
	}
}
