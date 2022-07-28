package webhookbuilder_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	webhookbuilder "github.com/sourcegraph/sourcegraph/internal/repos/worker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestJobQueue(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	workerBaseStore := basestore.NewWithHandle(db.Handle())

	extSvcKind := "GITHUB"

	t.Run(extSvcKind, func(t *testing.T) {
		recordID := 0
		job, err := dequeueJob(ctx, workerBaseStore, recordID)
		if err != nil && err.Error() != "expected 1 job to dequeue, found 0" {
			t.Fatal(err)
		}
		assertEqual(t, nil, nil, job)

		firstJob := &webhookbuilder.Job{
			RepoID:     1,
			RepoName:   "repo 1",
			ExtSvcKind: extSvcKind,
		}

		firstJobID, err := webhookbuilder.EnqueueJob(ctx, workerBaseStore, firstJob)
		if err != nil {
			t.Fatal(err)
		}

		secondJob := &webhookbuilder.Job{
			RepoID:     1,
			RepoName:   "repo 2",
			ExtSvcKind: extSvcKind,
		}

		secondJobID, err := webhookbuilder.EnqueueJob(ctx, workerBaseStore, secondJob)
		if err != nil {
			t.Fatal(err)
		}

		firstDequeuedJob, err := dequeueJob(ctx, workerBaseStore, firstJobID)
		assertEqual(t, err, firstJob, firstDequeuedJob)

		secondDequeuedJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
		assertEqual(t, err, secondJob, secondDequeuedJob)
	})
}

func dequeueJob(ctx context.Context, workerBaseStore *basestore.Store, recordID int) (*webhookbuilder.Job, error) {
	tx, err := workerBaseStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	rows, err := tx.Query(ctx, sqlf.Sprintf(dequeueJobFmtStr, recordID))
	if err != nil {
		return nil, err
	}

	jobs, err := doScanWebhookBuildJobs(rows, nil)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 1 {
		return nil, errors.Newf("expected 1 job to dequeue, found %v", len(jobs))
	}

	return jobs[0], nil
}

const dequeueJobFmtStr = `
-- source: internal/repos/worker/worker.go:dequeueJob
SELECT
	repo_id,
	repo_name,
	extsvc_kind,
	queued_at,
	id,
	state,
	failure_message,
	started_at,
	finished_at,
	process_after,
	num_resets,
	num_failures,
	execution_logs
FROM webhook_build_jobs
WHERE id = %s
`

func doScanWebhookBuildJobs(rows *sql.Rows, err error) ([]*webhookbuilder.Job, error) {
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var jobs []*webhookbuilder.Job

	for rows.Next() {
		var job webhookbuilder.Job
		if err := rows.Scan(
			// Webhook builder fields
			&job.RepoID,
			&job.RepoName,
			&job.ExtSvcKind,
			&job.QueuedAt,

			// Standard dbworker fields
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			pq.Array(&job.ExecutionLogs),
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func assertEqual(t *testing.T, err error, want *webhookbuilder.Job, have *webhookbuilder.Job) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	if have == nil {
		if have != nil {
			t.Fatal(errors.New("expected nil job, got non-nil job"))
		}
		return
	}

	if want.RepoID != have.RepoID ||
		want.RepoName != have.RepoName ||
		want.ExtSvcKind != have.ExtSvcKind {
		t.Fatal(errors.New("have, want not the same"))
	}
}
