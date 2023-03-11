package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestNumRepositoriesWithCodeIntelligence(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	insertUploads(t, db,
		Upload{ID: 100, RepositoryID: 50},
		Upload{ID: 101, RepositoryID: 51},
		Upload{ID: 102, RepositoryID: 52}, // Not in commit graph
		Upload{ID: 103, RepositoryID: 53}, // Not on default branch
	)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO lsif_uploads_visible_at_tip
			(repository_id, upload_id, is_default_branch)
		VALUES
			(50, 100, true),
			(51, 101, true),
			(53, 103, false)
	`); err != nil {
		t.Fatalf("unexpected error inserting visible uploads: %s", err)
	}

	count, err := store.NumRepositoriesWithCodeIntelligence(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting top repositories to configure: %s", err)
	}
	if expected := 2; count != expected {
		t.Fatalf("unexpected number of repositories. want=%d have=%d", expected, count)
	}
}

func TestRepositoryIDsWithErrors(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	now := time.Now()
	t1 := now.Add(-time.Minute * 1)
	t2 := now.Add(-time.Minute * 2)
	t3 := now.Add(-time.Minute * 3)

	insertUploads(t, db,
		Upload{ID: 100, RepositoryID: 50},                  // Repo 50 = success (no index)
		Upload{ID: 101, RepositoryID: 51},                  // Repo 51 = success (+ successful index)
		Upload{ID: 103, RepositoryID: 53, State: "failed"}, // Repo 53 = failed

		// Repo 54 = multiple failures for same project
		Upload{ID: 150, RepositoryID: 54, State: "failed", FinishedAt: &t1},
		Upload{ID: 151, RepositoryID: 54, State: "failed", FinishedAt: &t2},
		Upload{ID: 152, RepositoryID: 54, State: "failed", FinishedAt: &t3},

		// Repo 55 = multiple failures for different projects
		Upload{ID: 160, RepositoryID: 55, State: "failed", FinishedAt: &t1, Root: "proj1"},
		Upload{ID: 161, RepositoryID: 55, State: "failed", FinishedAt: &t2, Root: "proj2"},
		Upload{ID: 162, RepositoryID: 55, State: "failed", FinishedAt: &t3, Root: "proj3"},

		// Repo 58 = multiple failures with later success (not counted)
		Upload{ID: 170, RepositoryID: 58, State: "completed", FinishedAt: &t1},
		Upload{ID: 171, RepositoryID: 58, State: "failed", FinishedAt: &t2},
		Upload{ID: 172, RepositoryID: 58, State: "failed", FinishedAt: &t3},
	)
	insertIndexes(t, db,
		types.Index{ID: 201, RepositoryID: 51},                  // Repo 51 = success
		types.Index{ID: 202, RepositoryID: 52, State: "failed"}, // Repo 52 = failing index
		types.Index{ID: 203, RepositoryID: 53},                  // Repo 53 = success (+ failing upload)

		// Repo 56 = multiple failures for same project
		types.Index{ID: 250, RepositoryID: 56, State: "failed", FinishedAt: &t1},
		types.Index{ID: 251, RepositoryID: 56, State: "failed", FinishedAt: &t2},
		types.Index{ID: 252, RepositoryID: 56, State: "failed", FinishedAt: &t3},

		// Repo 57 = multiple failures for different projects
		types.Index{ID: 260, RepositoryID: 57, State: "failed", FinishedAt: &t1, Root: "proj1"},
		types.Index{ID: 261, RepositoryID: 57, State: "failed", FinishedAt: &t2, Root: "proj2"},
		types.Index{ID: 262, RepositoryID: 57, State: "failed", FinishedAt: &t3, Root: "proj3"},
	)

	// Query page 1
	repositoriesWithCount, totalCount, err := store.RepositoryIDsWithErrors(ctx, 0, 4)
	if err != nil {
		t.Fatalf("unexpected error getting repositories with errors: %s", err)
	}
	if expected := 6; totalCount != expected {
		t.Fatalf("unexpected total number of repositories. want=%d have=%d", expected, totalCount)
	}
	expected := []shared.RepositoryWithCount{
		{RepositoryID: 55, Count: 3},
		{RepositoryID: 57, Count: 3},
		{RepositoryID: 52, Count: 1},
		{RepositoryID: 53, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}

	// Query page 2
	repositoriesWithCount, _, err = store.RepositoryIDsWithErrors(ctx, 4, 4)
	if err != nil {
		t.Fatalf("unexpected error getting repositories with errors: %s", err)
	}
	expected = []shared.RepositoryWithCount{
		{RepositoryID: 54, Count: 1},
		{RepositoryID: 56, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}
}

func TestRepositoryIDsWithConfiguration(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	testIndexerList := map[string]shared.AvailableIndexer{
		"test-indexer": {
			Roots: []string{"proj1", "proj2", "proj3"},
			Indexer: types.CodeIntelIndexer{
				Name: "test-indexer",
			},
		},
	}

	for i := 0; i < 20; i++ {
		insertRepo(t, db, 50+i, fmt.Sprintf("test%d", i))

		if err := store.SetConfigurationSummary(ctx, 50+i, i*300, testIndexerList); err != nil {
			t.Fatalf("unexpected error setting configuration summary: %s", err)
		}
	}

	if err := store.TruncateConfigurationSummary(ctx, 10); err != nil {
		t.Fatalf("unexpected error truncating configuration summary: %s", err)
	}

	repositoriesWithCount, totalCount, err := store.RepositoryIDsWithConfiguration(ctx, 0, 5)
	if err != nil {
		t.Fatalf("unexpected error getting repositories with configuration: %s", err)
	}
	if expected := 10; totalCount != expected {
		t.Fatalf("unexpected total number of repositories. want=%d have=%d", expected, totalCount)
	}
	expected := []shared.RepositoryWithAvailableIndexers{
		{RepositoryID: 69, AvailableIndexers: testIndexerList},
		{RepositoryID: 68, AvailableIndexers: testIndexerList},
		{RepositoryID: 67, AvailableIndexers: testIndexerList},
		{RepositoryID: 66, AvailableIndexers: testIndexerList},
		{RepositoryID: 65, AvailableIndexers: testIndexerList},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}
}

func TestTopRepositoriesToConfigure(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	insertEvent := func(name string, repositoryID int, maxAge time.Duration) {
		query := `
			INSERT INTO event_logs (name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
			VALUES ($1, $2, '', 0, 'internal', 'test', 'dev', NOW() - ($3 * '1 hour'::interval))
		`
		if _, err := db.ExecContext(ctx, query, name, fmt.Sprintf(`{"repositoryId": %d}`, repositoryID), int(maxAge/time.Hour)); err != nil {
			t.Fatalf("unexpected error inserting events: %s", err)
		}
	}

	for i := 0; i < 50; i++ {
		insertRepo(t, db, 50+i, fmt.Sprintf("test%d", i))
	}
	for i := 0; i < 10; i++ {
		insertEvent("codeintel.searchHover", 60+i%3, 1)
	}
	for j := 0; j < 10; j++ {
		insertEvent("codeintel.searchHover", 70+j, 1)
	}

	insertEvent("codeintel.searchDefinitions", 50, 1)
	insertEvent("codeintel.searchDefinitions", 50, 1)
	insertEvent("codeintel.searchDefinitions.xrepo", 50, 1)
	insertEvent("search.symbol", 50, 1)                               // unmatched name
	insertEvent("codeintel.searchDefinitions", 50, eventLogsWindow*2) // out of window

	repositoriesWithCount, err := store.TopRepositoriesToConfigure(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error getting top repositories to configure: %s", err)
	}
	expected := []shared.RepositoryWithCount{
		{RepositoryID: 60, Count: 4}, // i=0,3,6,9
		{RepositoryID: 50, Count: 3}, // manual
		{RepositoryID: 61, Count: 3}, // i=1,4,7
		{RepositoryID: 62, Count: 3}, // i=2,5,8
		{RepositoryID: 70, Count: 1}, // j=0
		{RepositoryID: 71, Count: 1}, // j=1
		{RepositoryID: 72, Count: 1}, // j=2
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}
}
