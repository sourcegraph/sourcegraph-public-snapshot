package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

type printableRank struct{ value *int }

func (r printableRank) String() string {
	if r.value == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *r.value)
}

type printableTime struct{ value *time.Time }

func (r printableTime) String() string {
	if r.value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *r.value)
}

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// getDumpVisibilities returns a map from dump identifiers to its visibility. Fails the test on error.
func getDumpVisibilities(t *testing.T, db *sql.DB) map[int]bool {
	visibilities, err := scanVisibilities(db.Query("SELECT id, visible_at_tip FROM lsif_dumps_with_repository_name"))
	if err != nil {
		t.Fatalf("unexpected error while scanning dump visibility: %s", err)
	}

	return visibilities
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t *testing.T, db *sql.DB, uploads ...Upload) {
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
				visible_at_tip,
				uploaded_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				repository_id,
				indexer,
				num_parts,
				uploaded_parts,
				upload_size
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.VisibleAtTip,
			upload.UploadedAt,
			upload.State,
			upload.FailureMessage,
			upload.StartedAt,
			upload.FinishedAt,
			upload.ProcessAfter,
			upload.NumResets,
			upload.RepositoryID,
			upload.Indexer,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

// insertIndexes populates the lsif_indexes table with the given index models.
func insertIndexes(t *testing.T, db *sql.DB, indexes ...Index) {
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
				repository_id
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
			index.RepositoryID,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting index: %s", err)
		}
	}
}

// insertRepo creates a repository record with the given id and name. If there is already a repository
// with the given identifier, nothing happens
func insertRepo(t *testing.T, db *sql.DB, id int, name string) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}
}

// insertPackageReferences populates the lsif_references table with the given package references.
func insertPackageReferences(t *testing.T, store Store, packageReferences []types.PackageReference) {
	if err := store.UpdatePackageReferences(context.Background(), packageReferences); err != nil {
		t.Fatalf("unexpected error updating package references: %s", err)
	}
}

var UploadMetaComparer = cmp.Comparer(func(x, y UploadMeta) bool {
	return x.UploadID == y.UploadID && x.Distance == y.Distance
})

func scanVisibleUploads(rows *sql.Rows, queryErr error) (_ map[string][]UploadMeta, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	uploadMeta := map[string][]UploadMeta{}
	for rows.Next() {
		var commit string
		var uploadID int
		var distance int
		if err := rows.Scan(&commit, &uploadID, &distance); err != nil {
			return nil, err
		}

		uploadMeta[commit] = append(uploadMeta[commit], UploadMeta{
			UploadID: uploadID,
			Distance: distance,
		})
	}

	return uploadMeta, nil
}

func getVisibleUploads(t *testing.T, db *sql.DB, repositoryID int) map[string][]UploadMeta {
	query := sqlf.Sprintf(
		`SELECT commit, upload_id, distance FROM lsif_nearest_uploads WHERE repository_id = %s ORDER BY upload_id`,
		repositoryID,
	)
	uploads, err := scanVisibleUploads(db.QueryContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...))
	if err != nil {
		t.Fatalf("unexpected error getting visible uploads: %s", err)
	}

	return uploads
}

func getUploadsVisibleAtTip(t *testing.T, db *sql.DB, repositoryID int) []int {
	query := sqlf.Sprintf(
		`SELECT upload_id FROM lsif_uploads_visible_at_tip WHERE repository_id = %s ORDER BY upload_id`,
		repositoryID,
	)

	ids, err := scanInts(db.QueryContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...))
	if err != nil {
		t.Fatalf("unexpected error getting uploads visible at tip: %s", err)
	}

	return ids
}

// unwrapStore gets the underlying store from a store interface value.
func unwrapStore(s Store) *store {
	if s, ok := s.(*store); ok {
		return s
	}

	if observed, ok := s.(*ObservedStore); ok {
		return unwrapStore(observed.store)
	}

	return nil
}
