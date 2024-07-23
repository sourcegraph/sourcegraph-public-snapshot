package dependencies

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sqlf "github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/prometheus/statsd_exporter/pkg/clock"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func Test_AutoIndexingManualEnqueuedDequeueOrder(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	raw := dbtest.NewDB(t)
	db := database.NewDB(logtest.Scoped(t), raw)

	opts := IndexWorkerStoreOptions
	workerstore := store.New(observation.TestContextTB(t), db.Handle(), opts)

	for i, test := range []struct {
		jobs   []shared.AutoIndexJob
		nextID int
	}{
		{
			jobs: []shared.AutoIndexJob{
				{ID: 1, RepositoryID: 1, EnqueuerUserID: 51234},
				{ID: 2, RepositoryID: 4},
			},
			nextID: 1,
		},
		{
			jobs: []shared.AutoIndexJob{
				{ID: 1, RepositoryID: 1, EnqueuerUserID: 50, State: "completed", FinishedAt: dbutil.NullTimeColumn(clock.Now().Add(-time.Hour * 3))},
				{ID: 2, RepositoryID: 2},
				{ID: 3, RepositoryID: 1, EnqueuerUserID: 1},
			},
			nextID: 3,
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if _, err := db.ExecContext(context.Background(), "TRUNCATE lsif_indexes RESTART IDENTITY CASCADE"); err != nil {
				t.Fatal(err)
			}
			insertAutoIndexJobs(t, db, test.jobs...)
			job, _, err := workerstore.Dequeue(context.Background(), "borgir", nil)
			if err != nil {
				t.Fatal(err)
			}

			if job.ID != test.nextID {
				t.Fatalf("unexpected next index job candidate (got=%d,want=%d)", job.ID, test.nextID)
			}
		})
	}
}

func insertAutoIndexJobs(t testing.TB, db database.DB, jobs ...shared.AutoIndexJob) {
	for _, job := range jobs {
		if job.Commit == "" {
			job.Commit = fmt.Sprintf("%040d", job.ID)
		}
		if job.State == "" {
			job.State = "queued"
		}
		if job.RepositoryID == 0 {
			job.RepositoryID = 50
		}
		if job.DockerSteps == nil {
			job.DockerSteps = []shared.DockerStep{}
		}
		if job.IndexerArgs == nil {
			job.IndexerArgs = []string{}
		}
		if job.LocalSteps == nil {
			job.LocalSteps = []string{}
		}

		// Ensure we have a repo for the inner join in select queries
		insertRepo(t, db, job.RepositoryID, job.RepositoryName)

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
				should_reindex,
				enqueuer_user_id
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			job.ID,
			job.Commit,
			job.QueuedAt,
			job.State,
			job.FailureMessage,
			job.StartedAt,
			job.FinishedAt,
			job.ProcessAfter,
			job.NumResets,
			job.NumFailures,
			job.RepositoryID,
			pq.Array(job.DockerSteps),
			job.Root,
			job.Indexer,
			pq.Array(job.IndexerArgs),
			job.Outfile,
			pq.Array(job.ExecutionLogs),
			pq.Array(job.LocalSteps),
			job.ShouldReindex,
			job.EnqueuerUserID,
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
