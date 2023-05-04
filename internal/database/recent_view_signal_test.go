package database

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Creating 3 repos.
	err = db.Repos().Create(ctx, &types.Repo{ID: 1, Name: "github.com/sourcegraph/sourcegraph"}, &types.Repo{ID: 2, Name: "github.com/sourcegraph/sourcegraph2"}, &types.Repo{ID: 3, Name: "github.com/sourcegraph/sourcegraph3"})
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
		assert.Equal(t, 100, summaries[0].ViewsCount)
		clearTable(ctx, db)
	})
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
	_, err = db.QueryContext(ctx, "INSERT INTO repo_paths (repo_id, absolute_path, parent_id) VALUES (1, '', NULL), (1, 'src', 1), (1, 'src/abc', 2), (1, 'src/cde', 2)")
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	err = store.InsertPaths(ctx, 1, map[int]int{1: 10, 2: 100, 3: 1000})
	require.NoError(t, err)
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)
	assert.Len(t, summaries, 3)
	for _, summary := range summaries {
		assert.Equal(t, int(math.Pow(float64(10), float64(summary.FilePathID))), summary.ViewsCount)
	}
}

func TestRecentViewSignalStore_InsertPaths_HardLimit(t *testing.T) {
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
	inserts := []*sqlf.Query{sqlf.Sprintf("(1, '', NULL)"), sqlf.Sprintf("(1, 'src', 1)"), sqlf.Sprintf("(1, 'src/abc', 2)"), sqlf.Sprintf("(1, 'src/cde', 2)")}
	for i := 0; i < 5496; i++ {
		inserts = append(inserts, sqlf.Sprintf("(1, %s, 1)", strconv.Itoa(i)))
	}
	query := sqlf.Sprintf("INSERT INTO repo_paths (repo_id, absolute_path, parent_id) VALUES %s", sqlf.Join(inserts, ","))
	_, err = db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	require.NoError(t, err)

	store := RecentViewSignalStoreWith(db, logger)

	paths := make(map[int]int)
	for i := 1; i < 5500; i++ {
		paths[i] = i
	}

	// Inserting 5499 paths, but only 5000 should be inserted.
	err = store.InsertPaths(ctx, 1, paths)
	require.NoError(t, err)
	summaries, err := store.List(ctx, ListRecentViewSignalOpts{})
	require.NoError(t, err)
	require.Len(t, summaries, 5000)
}
