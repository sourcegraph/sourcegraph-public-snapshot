package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// getDumpVisibilities returns a map from dump identifiers to its visibility. Fails the test on error.
func getDumpVisibilities(t *testing.T, db *sql.DB) map[int]bool {
	visibilities, err := scanVisibilities(db.Query("SELECT id, visible_at_tip FROM lsif_dumps"))
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

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				id,
				commit,
				root,
				visible_at_tip,
				uploaded_at,
				state,
				failure_summary,
				failure_stacktrace,
				started_at,
				finished_at,
				tracing_context,
				repository_id,
				indexer
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.VisibleAtTip,
			upload.UploadedAt,
			upload.State,
			upload.FailureSummary,
			upload.FailureStacktrace,
			upload.StartedAt,
			upload.FinishedAt,
			upload.TracingContext,
			upload.RepositoryID,
			upload.Indexer,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting dump: %s", err)
		}
	}
}

// insertPackageReferences populates the lsif_references table with the given package references.
func insertPackageReferences(t *testing.T, db *dbImpl, packageReferences []types.PackageReference) {
	if err := db.UpdatePackageReferences(context.Background(), packageReferences); err != nil {
		t.Fatalf("unexpected error updating package references: %s", err)
	}
}
