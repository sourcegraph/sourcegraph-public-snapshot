package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
	db := NewDB(logger, dbtest.NewDB(t))
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
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
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
	db := NewDB(logger, dbtest.NewDB(t))
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
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
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
	db := NewDB(logger, dbtest.NewDB(t))
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
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
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
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
		require.NoError(t, err)
		assert.Len(t, summaries, 2)
		assert.Equal(t, 3, summaries[0].FilePathID)
		assert.Equal(t, 20, summaries[0].ViewsCount)
		assert.Equal(t, 2, summaries[1].FilePathID)
		assert.Equal(t, 10, summaries[1].ViewsCount)
		clearTable(ctx, db)
	})

	t.Run("inserting conflicting entry will update it", func(t *testing.T) {
		err = store.Insert(ctx, 1, 2, 10)
		require.NoError(t, err)
		summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		assert.Equal(t, 2, summaries[0].FilePathID)
		assert.Equal(t, 10, summaries[0].ViewsCount)

		// Inserting a conflicting entry.
		err = store.Insert(ctx, 1, 2, 100)
		require.NoError(t, err)
		summaries, err = store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
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
	db := NewDB(logger, dbtest.NewDB(t))
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
	got, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
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
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user.
	_, err := db.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)

	// Creating a repo.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating 15 paths.
	var paths []string
	for i := 1; i <= 15; i++ {
		paths = append(paths, fmt.Sprintf("src/file%d", i))
	}
	pathIDs, err := ensureRepoPaths(ctx, storeFrom(t, db), paths, 1)
	require.NoError(t, err)

	store := &recentViewSignalStore{Store: basestore.NewWithHandle(db.Handle()), Logger: logger}

	counts := map[int]int{}
	for _, id := range pathIDs {
		counts[id] = 10
	}

	err = store.insertPaths(ctx, 1, counts, 10) // batch size of 10 is smaller than the total 15 paths
	require.NoError(t, err)
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{IncludeAllPaths: true})
	require.NoError(t, err)
	require.Len(t, summaries, 17) // Two extra entries - repo root and 'src' directory
}

func TestRecentViewSignalStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	d := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating 2 users.
	user1, err := d.Users().Create(ctx, NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := d.Users().Create(ctx, NewUser{Username: "user2"})
	require.NoError(t, err)

	// Creating a repo.
	var repoID api.RepoID = 1
	err = d.Repos().Create(ctx, &types.Repo{ID: repoID, Name: "github.com/sourcegraph/sourcegraph"})
	require.NoError(t, err)

	// Creating some paths.
	paths := []string{"", "src", "src/abc", "src/cde", "src/def"}
	ids, err := ensureRepoPaths(ctx, d.(*db).Store, paths, repoID)
	require.NoError(t, err)
	pathIDs := map[string]int{}
	for i, p := range paths {
		pathIDs[p] = ids[i]
	}

	viewCounts1 := map[string]int{
		"":        10000,
		"src":     1000,
		"src/abc": 100,
		"src/cde": 10, // different path than in viewCounts2
	}
	viewCounts2 := map[string]int{
		"":        20000,
		"src":     2000,
		"src/abc": 200,
		"src/def": 20, // different path than in viewCounts1
	}
	for path, count := range viewCounts1 {
		require.NoError(t, d.RecentViewSignal().Insert(ctx, user1.ID, pathIDs[path], count))
	}
	for path, count := range viewCounts2 {
		require.NoError(t, d.RecentViewSignal().Insert(ctx, user2.ID, pathIDs[path], count))
	}

	// As IDs of signals aren't returned, we can rely on counts because of strict
	// mapping.
	testCases := map[string]struct {
		opts              ListRecentViewSignalOpts
		expectedCounts    []int
		expectedNoEntries bool
	}{
		"list values for the whole table": {
			opts:           ListRecentViewSignalOpts{IncludeAllPaths: true},
			expectedCounts: []int{20000, 10000, 2000, 1000, 200, 100, 20, 10},
		},
		"list values for root path": {
			opts:           ListRecentViewSignalOpts{},
			expectedCounts: []int{viewCounts2[""], viewCounts1[""]},
		},
		"list values for root path with min threashold": {
			opts:           ListRecentViewSignalOpts{MinThreshold: 15000},
			expectedCounts: []int{viewCounts2[""]},
		},
		"filter by viewer ID": {
			opts:           ListRecentViewSignalOpts{ViewerUserID: 1},
			expectedCounts: []int{viewCounts1[""]},
		},
		"filter by viewer ID which isn't present": {
			opts:              ListRecentViewSignalOpts{ViewerUserID: -1},
			expectedNoEntries: true,
		},
		"filter by repo ID": {
			opts:           ListRecentViewSignalOpts{RepoID: 1},
			expectedCounts: []int{viewCounts2[""], viewCounts1[""]},
		},
		"filter by repo ID which isn't present": {
			opts:              ListRecentViewSignalOpts{RepoID: 2},
			expectedNoEntries: true,
		},
		"filter by path": {
			opts:           ListRecentViewSignalOpts{Path: "src/cde"},
			expectedCounts: []int{viewCounts1["src/cde"]},
		},
		"filter by path which isn't present": {
			opts:              ListRecentViewSignalOpts{Path: "lol"},
			expectedNoEntries: true,
		},
		"limit, offset": {
			opts:           ListRecentViewSignalOpts{LimitOffset: &LimitOffset{Limit: 1, Offset: 1}},
			expectedCounts: []int{viewCounts1[""]},
		},
		"limit": {
			opts:           ListRecentViewSignalOpts{LimitOffset: &LimitOffset{Limit: 1}},
			expectedCounts: []int{viewCounts2[""]},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			gotSummaries, err := d.RecentViewSignal().List(ctx, test.opts)
			require.NoError(t, err)
			if test.expectedNoEntries {
				assert.Empty(t, gotSummaries)
				return
			}
			var gotCounts []int
			for _, s := range gotSummaries {
				gotCounts = append(gotCounts, s.ViewsCount)
			}
			assert.Equal(t, test.expectedCounts, gotCounts)
		})
	}
}
