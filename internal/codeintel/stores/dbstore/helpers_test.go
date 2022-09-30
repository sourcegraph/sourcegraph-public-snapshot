package dbstore

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t testing.TB, db database.DB, uploads ...Upload) {
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
