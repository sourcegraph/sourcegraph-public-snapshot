package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetStaleSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		shared.Index{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		shared.Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		shared.Index{ID: 3, RepositoryID: 50, Commit: makeCommit(3)},
		shared.Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		shared.Index{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
	)

	sourcedCommits, err := store.GetStaleSourcedCommits(context.Background(), time.Minute, 5, now)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits := []shared.SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1), makeCommit(2), makeCommit(3)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(6)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// 120s away from next check (threshold is 60s)
	if _, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(1), now); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	// 30s away from next check (threshold is 60s)
	if _, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(2), now.Add(time.Second*90)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	sourcedCommits, err = store.GetStaleSourcedCommits(context.Background(), time.Minute, 5, now.Add(time.Minute*2))
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits = []shared.SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1), makeCommit(3)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(6)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}

func TestUpdateSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		shared.Index{ID: 1, RepositoryID: 50, Commit: makeCommit(3)},
		shared.Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		shared.Index{ID: 3, RepositoryID: 52, Commit: makeCommit(7)},
		shared.Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		shared.Index{ID: 5, RepositoryID: 50, Commit: makeCommit(1)},
	)

	indexesUpdated, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(1), now)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if indexesUpdated != 1 {
		t.Fatalf("unexpected indexes updated. want=%d have=%d", 1, indexesUpdated)
	}

	indexStates, err := getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	expectedIndexStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}
}

func TestDeleteSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	insertIndexes(t, db,
		shared.Index{ID: 1, RepositoryID: 50, Commit: makeCommit(3)},
		shared.Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		shared.Index{ID: 3, RepositoryID: 52, Commit: makeCommit(7)},
		shared.Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		shared.Index{ID: 5, RepositoryID: 50, Commit: makeCommit(1)},
	)

	indexesDeleted, err := store.DeleteSourcedCommits(context.Background(), 52, makeCommit(7), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if indexesDeleted != 1 {
		t.Fatalf("unexpected number of indexes deleted. want=%d have=%d", 1, indexesDeleted)
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
}

// insertIndexes populates the lsif_indexes table with the given index models.
func insertIndexes(t testing.TB, db database.DB, indexes ...shared.Index) {
	for _, index := range indexes {
		if index.Commit == "" {
			index.Commit = makeCommit(index.ID)
		}
		if index.State == "" {
			index.State = "completed"
		}
		if index.RepositoryID == 0 {
			index.RepositoryID = 50
		}
		if index.DockerSteps == nil {
			index.DockerSteps = []shared.DockerStep{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}
		if index.LocalSteps == nil {
			index.LocalSteps = []string{}
		}

		// Ensure we have a repo for the inner join in select queries
		insertRepo(t, db, index.RepositoryID, index.RepositoryName)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_indexes (
				id,
				commit,
				queued_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				num_failures,
				repository_id,
				docker_steps,
				root,
				indexer,
				indexer_args,
				outfile,
				execution_logs,
				local_steps
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			index.ID,
			index.Commit,
			index.QueuedAt,
			index.State,
			index.FailureMessage,
			index.StartedAt,
			index.FinishedAt,
			index.ProcessAfter,
			index.NumResets,
			index.NumFailures,
			index.RepositoryID,
			pq.Array(index.DockerSteps),
			index.Root,
			index.Indexer,
			pq.Array(index.IndexerArgs),
			index.Outfile,
			pq.Array(index.ExecutionLogs),
			pq.Array(index.LocalSteps),
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting index: %s", err)
		}
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

// scanStates scans pairs of id/states from the return value of `*Store.query`.
func scanStates(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	states := map[int]string{}
	for rows.Next() {
		var id int
		var state string
		if err := rows.Scan(&id, &state); err != nil {
			return nil, err
		}

		states[id] = strings.ToLower(state)
	}

	return states, nil
}

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// insertRepo creates a repository record with the given id and name. If there is already a repository
// with the given identifier, nothing happens
func insertRepo(t testing.TB, db database.DB, id int, name string) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HasPrefix(name, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}
}

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}
