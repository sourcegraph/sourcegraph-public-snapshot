package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func newTest(db database.DB) *store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("autoindexing.store", ""),
		operations: newOperations(&observation.TestContext),
	}
}

func TestProcessStaleSourcedCommits(t *testing.T) {
	log := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(log, t)
	db := database.NewDB(log, sqlDB)
	store := newTest(db)

	ctx := context.Background()
	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		types.Index{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		types.Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		types.Index{ID: 3, RepositoryID: 50, Commit: makeCommit(3)},
		types.Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		types.Index{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
	)

	const (
		minimumTimeSinceLastCheck = time.Minute
		commitResolverBatchSize   = 5
	)

	// First update
	deleteCommit3 := func(ctx context.Context, repositoryID int, respositoryName, commit string) (bool, error) {
		return commit == makeCommit(3), nil
	}
	if _, numDeleted, err := store.processStaleSourcedCommits(
		ctx,
		minimumTimeSinceLastCheck,
		commitResolverBatchSize,
		deleteCommit3,
		now,
	); err != nil {
		t.Fatalf("unexpected error processing stale sourced commits: %s", err)
	} else if numDeleted != 1 {
		t.Fatalf("unexpected number of deleted indexes. want=%d have=%d", 1, numDeleted)
	}
	indexStates, err := getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	expectedIndexStates := map[int]string{
		1: "completed",
		2: "completed",
		// 3 was deleted
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}

	// Too soon after last update
	deleteCommit2 := func(ctx context.Context, repositoryID int, respositoryName, commit string) (bool, error) {
		return commit == makeCommit(2), nil
	}
	if _, numDeleted, err := store.processStaleSourcedCommits(
		ctx,
		minimumTimeSinceLastCheck,
		commitResolverBatchSize,
		deleteCommit2,
		now.Add(minimumTimeSinceLastCheck/2),
	); err != nil {
		t.Fatalf("unexpected error processing stale sourced commits: %s", err)
	} else if numDeleted != 0 {
		t.Fatalf("unexpected number of deleted indexes. want=%d have=%d", 0, numDeleted)
	}
	indexStates, err = getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	// no change in expectedIndexStates
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}

	// Enough time after previous update(s)
	if _, numDeleted, err := store.processStaleSourcedCommits(
		ctx,
		minimumTimeSinceLastCheck,
		commitResolverBatchSize,
		deleteCommit2,
		now.Add(minimumTimeSinceLastCheck/2*3),
	); err != nil {
		t.Fatalf("unexpected error processing stale sourced commits: %s", err)
	} else if numDeleted != 1 {
		t.Fatalf("unexpected number of deleted indexes. want=%d have=%d", 1, numDeleted)
	}
	indexStates, err = getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	expectedIndexStates = map[int]string{
		1: "completed",
		// 2 was deleted
		// 3 was deleted
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}
}

func getIndexStates(db database.DB, ids ...int) (map[int]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(
		`SELECT id, state FROM lsif_indexes WHERE id IN (%s)`,
		sqlf.Join(intsToQueries(ids), ", "),
	)

	return scanStates(db.QueryContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...))
}
