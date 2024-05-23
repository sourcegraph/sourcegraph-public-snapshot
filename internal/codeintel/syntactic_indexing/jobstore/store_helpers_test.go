package jobstore

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal/testutils"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/stretchr/testify/require"
)

func insertIndexRecords(t testing.TB, db database.DB, records ...SyntacticIndexingJob) {
	for _, index := range records {
		if index.Commit == "" {
			index.Commit = testutils.MakeCommit(index.ID)
		}
		if index.State == "" {
			index.State = Completed
		}
		if index.RepositoryID == 0 {
			index.RepositoryID = 50
		}
		// Ensure we have a repo for the inner join in select queries
		testutils.InsertRepo(t, db, index.RepositoryID, index.RepositoryName)

		query := sqlf.Sprintf(`
			INSERT INTO syntactic_scip_indexing_jobs (
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
				should_reindex,
				enqueuer_user_id
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
			index.ShouldReindex,
			index.EnqueuerUserID,
		)

		_, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...)
		require.NoError(t, err, "unexpected error while inserting index")
	}
}
