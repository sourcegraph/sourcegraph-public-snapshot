package store

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestSelectRepositoriesForIndexScan(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	now := timeutil.Now()
	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			(101, 50, 'policy 1', 'GIT_TREE', 'ab/', null, true, 0, false, true,  0, false),
			(102, 51, 'policy 2', 'GIT_TREE', 'cd/', null, true, 0, false, true,  0, false),
			(103, 52, 'policy 3', 'GIT_TREE', 'ef/', null, true, 0, false, true,  0, false),
			(104, 53, 'policy 4', 'GIT_TREE', 'gh/', null, true, 0, false, true,  0, false),
			(105, 54, 'policy 5', 'GIT_TREE', 'gh/', null, true, 0, false, false, 0, false)
	`
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	// Can return null last_index_scan
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 2, now); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 20 minutes later, first two repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*20)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 30 minutes later, all repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*30)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 90 minutes later, all repositories are visible
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*90)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// Make new invisible repository
	insertRepo(t, db, 54, "r4")

	// 95 minutes later, new repository is not yet visible
	if repositoryIDs, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*95)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositoryIDs); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	query = `UPDATE lsif_configuration_policies SET indexing_enabled = true WHERE id = 105`
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	// 100 minutes later, only new repository is visible
	if repositoryIDs, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*100)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{54}, repositoryIDs); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}
}

func TestSelectRepositoriesForIndexScanWithGlobalPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	now := timeutil.Now()
	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			(101, NULL, 'policy 1', 'GIT_TREE', 'ab/', null, true, 0, false, true, 0, false)
	`
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	// Returns nothing when disabled
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, false, nil, 100, now); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// Returns at most configured limit
	limit := 2

	// Can return null last_index_scan
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, &limit, 100, now); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 20 minutes later, first two repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*20)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 30 minutes later, all repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*30)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 90 minutes later, all repositories are visible
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), "lsif_last_index_scan", "last_index_scan_at", time.Hour, true, nil, 100, now.Add(time.Minute*90)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}
}

func TestSelectRepositoriesForIndexScanInDifferentTable(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	now := timeutil.Now()
	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits
		) VALUES
			(101, NULL, 'policy 1', 'GIT_TREE', 'ab/', null, true, 0, false, true, 0, false)
	`
	if _, err := db.ExecContext(context.Background(), query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	// Create a new table
	query = `
		CREATE TABLE last_incredible_testing_scan (
			repository_id integer NOT NULL PRIMARY KEY,
			last_incredible_testing_scan_at timestamp with time zone NOT NULL
		)
	`

	if _, err := db.ExecContext(context.Background(), query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	tableName := "last_incredible_testing_scan"
	columnName := "last_incredible_testing_scan_at"

	// Returns at most configured limit
	limit := 2

	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), tableName, columnName, time.Hour, true, &limit, 100, now); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 20 minutes later, first two repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), tableName, columnName, time.Hour, true, nil, 100, now.Add(time.Minute*20)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 30 minutes later, all repositories are still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), tableName, columnName, time.Hour, true, nil, 100, now.Add(time.Minute*30)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 90 minutes later, all repositories are visible
	if repositories, err := store.GetRepositoriesForIndexScan(context.Background(), tableName, columnName, time.Hour, true, nil, 100, now.Add(time.Minute*90)); err != nil {
		t.Fatalf("unexpected error fetching repositories for index scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}
}

func TestSetRepositoryAsDirty(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	for _, repositoryID := range []int{50, 51, 52, 51, 52} {
		if err := store.SetRepositoryAsDirty(context.Background(), repositoryID); err != nil {
			t.Errorf("unexpected error marking repository as dirty: %s", err)
		}
	}

	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{50, 51, 52}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
	}
}

func TestGetRepositoriesMaxStaleAge(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO lsif_dirty_repositories (
			repository_id,
			update_token,
			dirty_token,
			set_dirty_at
		)
		VALUES
			(50, 10, 10, NOW() - '45 minutes'::interval), -- not dirty
			(51, 20, 25, NOW() - '30 minutes'::interval), -- dirty
			(52, 30, 35, NOW() - '20 minutes'::interval), -- dirty
			(53, 40, 45, NOW() - '30 minutes'::interval); -- no associated repo
	`); err != nil {
		t.Fatalf("unexpected error marking repostiory as dirty: %s", err)
	}

	age, err := store.GetRepositoriesMaxStaleAge(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if age.Round(time.Second) != 30*time.Minute {
		t.Fatalf("unexpected max age. want=%s have=%s", 30*time.Minute, age)
	}
}

func TestHasRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	testCases := []struct {
		repositoryID int
		exists       bool
	}{
		{50, true},
		{51, false},
		{52, false},
	}

	insertUploads(t, db, shared.Upload{ID: 1, RepositoryID: 50})
	insertUploads(t, db, shared.Upload{ID: 2, RepositoryID: 51, State: "deleted"})

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryID=%d", testCase.repositoryID)

		t.Run(name, func(t *testing.T) {
			exists, err := store.HasRepository(context.Background(), testCase.repositoryID)
			if err != nil {
				t.Fatalf("unexpected error checking if repository exists: %s", err)
			}
			if exists != testCase.exists {
				t.Errorf("unexpected exists. want=%v have=%v", testCase.exists, exists)
			}
		})
	}
}

func TestSetRepositoriesForRetentionScan(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, State: "completed"},
		shared.Upload{ID: 2, RepositoryID: 51, State: "completed"},
		shared.Upload{ID: 3, RepositoryID: 52, State: "completed"},
		shared.Upload{ID: 4, RepositoryID: 53, State: "completed"},
		shared.Upload{ID: 5, RepositoryID: 54, State: "errored"},
		shared.Upload{ID: 6, RepositoryID: 54, State: "deleted"},
	)

	now := timeutil.Now()

	for _, repositoryID := range []int{50, 51, 52, 53, 54} {
		// Only call this to insert a record into the lsif_dirty_repositories table
		if err := store.SetRepositoryAsDirty(context.Background(), repositoryID); err != nil {
			t.Fatalf("unexpected error marking repository as dirty`: %s", err)
		}

		// Only call this to update the updated_at field in the lsif_dirty_repositories table
		if err := store.UpdateUploadsVisibleToCommits(context.Background(), repositoryID, gitdomain.ParseCommitGraph(nil), nil, time.Hour, time.Hour, 1, now); err != nil {
			t.Fatalf("unexpected error updating commit graph: %s", err)
		}
	}

	// Can return null last_index_scan
	if repositories, err := store.SetRepositoriesForRetentionScan(context.Background(), time.Hour, 2); err != nil {
		t.Fatalf("unexpected error fetching repositories for retention scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 20 minutes later, first two repositories are still on cooldown
	if repositories, err := store.SetRepositoriesForRetentionScanWithTime(context.Background(), time.Hour, 100, now.Add(time.Minute*20)); err != nil {
		t.Fatalf("unexpected error fetching repositories for retention scan: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 30 minutes later, all repositories are still on cooldown
	if repositories, err := store.SetRepositoriesForRetentionScanWithTime(context.Background(), time.Hour, 100, now.Add(time.Minute*30)); err != nil {
		t.Fatalf("unexpected error fetching repositories for retention scan: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// 90 minutes later, all repositories are visible
	if repositories, err := store.SetRepositoriesForRetentionScanWithTime(context.Background(), time.Hour, 100, now.Add(time.Minute*90)); err != nil {
		t.Fatalf("unexpected error fetching repositories for retention scan: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}

	// Make repository 5 newly visible
	if _, err := db.ExecContext(context.Background(), `UPDATE lsif_uploads SET state = 'completed' WHERE id = 5`); err != nil {
		t.Fatalf("unexpected error updating upload: %s", err)
	}

	// 95 minutes later, only new repository is visible
	if repositoryIDs, err := store.SetRepositoriesForRetentionScanWithTime(context.Background(), time.Hour, 100, now.Add(time.Minute*95)); err != nil {
		t.Fatalf("unexpected error fetching repositories for retention scan: %s", err)
	} else if diff := cmp.Diff([]int{54}, repositoryIDs); diff != "" {
		t.Fatalf("unexpected repository list (-want +got):\n%s", diff)
	}
}

func TestSkipsDeletedRepositories(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertRepo(t, db, 50, "should not be dirty")
	deleteRepo(t, db, 50, time.Now())

	insertRepo(t, db, 51, "should be dirty")

	// NOTE: We did not insert 52, so it should not show up as dirty, even though we mark it below.

	for _, repositoryID := range []int{50, 51, 52} {
		if err := store.SetRepositoryAsDirty(context.Background(), repositoryID); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{51}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
	}
}

// Marks a repo as deleted
func deleteRepo(t testing.TB, db database.DB, id int, deleted_at time.Time) {
	query := sqlf.Sprintf(
		`UPDATE repo SET deleted_at = %s WHERE id = %s`,
		deleted_at,
		id,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while deleting repository: %s", err)
	}
}

func testStoreWithoutConfigurationPolicies(t *testing.T, db database.DB) Store {
	if _, err := db.ExecContext(context.Background(), `TRUNCATE lsif_configuration_policies`); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	store := New(db, &observation.TestContext)
	return store
}
