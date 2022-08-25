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

	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
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

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		shared.Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		shared.Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		shared.Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		shared.Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
		shared.Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(8)},
	)

	sourcedCommits, err := store.GetStaleSourcedCommits(context.Background(), time.Minute, 5, now)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits := []shared.SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7), makeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// 120s away from next check (threshold is 60s)
	if _, err := store.UpdateSourcedCommits(context.Background(), 52, makeCommit(7), now); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	// 30s away from next check (threshold is 60s)
	if _, err := store.UpdateSourcedCommits(context.Background(), 52, makeCommit(8), now.Add(time.Second*90)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	sourcedCommits, err = store.GetStaleSourcedCommits(context.Background(), time.Minute, 5, now.Add(time.Minute*2))
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits = []shared.SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5)}},
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

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		shared.Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		shared.Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		shared.Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		shared.Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
		shared.Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(7), State: "uploading"},
	)

	uploadsUpdated, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(1), now)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if uploadsUpdated != 2 {
		t.Fatalf("unexpected uploads updated. want=%d have=%d", 2, uploadsUpdated)
	}

	uploadStates, err := getUploadStates(db, 1, 2, 3, 4, 5, 6)
	if err != nil {
		t.Fatalf("unexpected error fetching upload states: %s", err)
	}
	expectedUploadStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "completed",
		6: "uploading",
	}
	if diff := cmp.Diff(expectedUploadStates, uploadStates); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}
}

func TestDeleteSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		shared.Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		shared.Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		shared.Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		shared.Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
		shared.Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(7), State: "uploading", UploadedAt: now.Add(-time.Minute * 90)},
		shared.Upload{ID: 7, RepositoryID: 52, Commit: makeCommit(7), State: "queued", UploadedAt: now.Add(-time.Minute * 30)},
	)

	uploadsUpdated, uploadsDeleted, err := store.DeleteSourcedCommits(context.Background(), 52, makeCommit(7), time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if uploadsUpdated != 1 {
		t.Fatalf("unexpected number of uploads updated. want=%d have=%d", 1, uploadsUpdated)
	}
	if uploadsDeleted != 2 {
		t.Fatalf("unexpected number of uploads deleted. want=%d have=%d", 2, uploadsDeleted)
	}

	uploadStates, err := getUploadStates(db, 1, 2, 3, 4, 5, 6, 7)
	if err != nil {
		t.Fatalf("unexpected error fetching upload states: %s", err)
	}
	expectedUploadStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "deleting",
		6: "deleted",
		7: "queued",
	}
	if diff := cmp.Diff(expectedUploadStates, uploadStates); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}
}

func TestHasCommit(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	testCases := []struct {
		repositoryID int
		commit       string
		exists       bool
	}{
		{50, makeCommit(1), true},
		{50, makeCommit(2), false},
		{51, makeCommit(1), false},
	}

	insertNearestUploads(t, db, 50, map[string][]commitgraph.UploadMeta{makeCommit(1): {{UploadID: 42, Distance: 1}}})
	insertNearestUploads(t, db, 51, map[string][]commitgraph.UploadMeta{makeCommit(2): {{UploadID: 43, Distance: 2}}})

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryID=%d commit=%s", testCase.repositoryID, testCase.commit)

		t.Run(name, func(t *testing.T) {
			exists, err := store.HasCommit(context.Background(), testCase.repositoryID, testCase.commit)
			if err != nil {
				t.Fatalf("unexpected error checking if commit exists: %s", err)
			}
			if exists != testCase.exists {
				t.Errorf("unexpected exists. want=%v have=%v", testCase.exists, exists)
			}
		})
	}
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t testing.TB, db database.DB, uploads ...shared.Upload) {
	for _, upload := range uploads {
		if upload.Commit == "" {
			upload.Commit = makeCommit(upload.ID)
		}
		if upload.State == "" {
			upload.State = "completed"
		}
		if upload.RepositoryID == 0 {
			upload.RepositoryID = 50
		}
		if upload.Indexer == "" {
			upload.Indexer = "lsif-go"
		}
		if upload.IndexerVersion == "" {
			upload.IndexerVersion = "latest"
		}
		if upload.UploadedParts == nil {
			upload.UploadedParts = []int{}
		}

		// Ensure we have a repo for the inner join in select queries
		insertRepo(t, db, upload.RepositoryID, upload.RepositoryName)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				id,
				commit,
				root,
				uploaded_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				num_failures,
				repository_id,
				indexer,
				indexer_version,
				num_parts,
				uploaded_parts,
				upload_size,
				associated_index_id
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.UploadedAt,
			upload.State,
			upload.FailureMessage,
			upload.StartedAt,
			upload.FinishedAt,
			upload.ProcessAfter,
			upload.NumResets,
			upload.NumFailures,
			upload.RepositoryID,
			upload.Indexer,
			upload.IndexerVersion,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
			upload.AssociatedIndexID,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

func updateUploads(t testing.TB, db database.DB, uploads ...shared.Upload) {
	for _, upload := range uploads {
		query := sqlf.Sprintf(`
			UPDATE lsif_uploads
			SET
				commit = COALESCE(NULLIF(%s, ''), commit),
				root = COALESCE(NULLIF(%s, ''), root),
				uploaded_at = COALESCE(NULLIF(%s, '0001-01-01 00:00:00+00'::timestamptz), uploaded_at),
				state = COALESCE(NULLIF(%s, ''), state),
				failure_message  = COALESCE(%s, failure_message),
				started_at = COALESCE(%s, started_at),
				finished_at = COALESCE(%s, finished_at),
				process_after = COALESCE(%s, process_after),
				num_resets = COALESCE(NULLIF(%s, 0), num_resets),
				num_failures = COALESCE(NULLIF(%s, 0), num_failures),
				repository_id = COALESCE(NULLIF(%s, 0), repository_id),
				indexer = COALESCE(NULLIF(%s, ''), indexer),
				indexer_version = COALESCE(NULLIF(%s, ''), indexer_version),
				num_parts = COALESCE(NULLIF(%s, 0), num_parts),
				uploaded_parts = COALESCE(NULLIF(%s, '{}'::integer[]), uploaded_parts),
				upload_size = COALESCE(%s, upload_size),
				associated_index_id = COALESCE(%s, associated_index_id)
			WHERE id = %s
		`,
			upload.Commit,
			upload.Root,
			upload.UploadedAt,
			upload.State,
			upload.FailureMessage,
			upload.StartedAt,
			upload.FinishedAt,
			upload.ProcessAfter,
			upload.NumResets,
			upload.NumFailures,
			upload.RepositoryID,
			upload.Indexer,
			upload.IndexerVersion,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
			upload.AssociatedIndexID,
			upload.ID)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while updating upload: %s", err)
		}
	}
}

func deleteUploads(t testing.TB, db database.DB, uploads ...int) {
	for _, upload := range uploads {
		query := sqlf.Sprintf(`DELETE FROM lsif_uploads WHERE id = %s`, upload)
		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while deleting upload: %s", err)
		}
	}
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

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

func getUploadStates(db database.DB, ids ...int) (map[int]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(
		`SELECT id, state FROM lsif_uploads WHERE id IN (%s)`,
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
