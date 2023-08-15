package store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type upload struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	UncompressedSize  *int64
	Rank              *int
	AssociatedIndexID *int
	ShouldReindex     bool
}

func insertUploads(t testing.TB, db database.DB, uploads ...upload) {
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
				associated_index_id,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
			upload.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

func insertIndexes(t testing.TB, db database.DB, indexes ...uploadsshared.Index) {
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
			index.DockerSteps = []uploadsshared.DockerStep{}
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
				local_steps,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
			index.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting index: %s", err)
		}
	}
}

func insertRepo(t testing.TB, db database.DB, id int, name string) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HasPrefix(name, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at, private) VALUES (%s, %s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
		false,
	)
	if _, err := db.ExecContext(context.Background(), insertRepoQuery.Query(sqlf.PostgresBindVar), insertRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}

	status := "cloned"
	if strings.HasPrefix(name, "DELETED-") {
		status = "not_cloned"
	}
	updateGitserverRepoQuery := sqlf.Sprintf(
		`UPDATE gitserver_repos SET clone_status = %s WHERE repo_id = %s`,
		status,
		id,
	)
	if _, err := db.ExecContext(context.Background(), updateGitserverRepoQuery.Query(sqlf.PostgresBindVar), updateGitserverRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting gitserver repository: %s", err)
	}
}

func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
