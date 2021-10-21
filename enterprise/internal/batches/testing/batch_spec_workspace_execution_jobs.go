package testing

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type execStore interface {
	Exec(ctx context.Context, query *sqlf.Query) error
}

func UpdateJobState(t *testing.T, ctx context.Context, s execStore, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()

	const fmtStr = `
UPDATE batch_spec_workspace_execution_jobs
SET
	state = %s,
	started_at = %s,
	finished_at = %s,
	cancel = %s,
	worker_hostname = %s
WHERE
	id = %s
`

	q := sqlf.Sprintf(
		fmtStr,
		job.State,
		nullTimeColumn(job.StartedAt),
		nullTimeColumn(job.FinishedAt),
		job.Cancel,
		job.WorkerHostname,
		job.ID,
	)
	if err := s.Exec(ctx, q); err != nil {
		t.Fatal(err)
	}
}

type createBatchSpecWorkspaceExecutionJobStore interface {
	basestore.ShareableStore
	Clock() func() time.Time
}

type workspaceExecutionScanner = func(wj *btypes.BatchSpecWorkspaceExecutionJob, s dbutil.Scanner) error

func CreateBatchSpecWorkspaceExecutionJob(ctx context.Context, s createBatchSpecWorkspaceExecutionJobStore, scanFn workspaceExecutionScanner, jobs ...*btypes.BatchSpecWorkspaceExecutionJob) (err error) {
	inserter := func(inserter *batch.Inserter) error {
		for _, job := range jobs {
			if job.CreatedAt.IsZero() {
				job.CreatedAt = s.Clock()()
			}

			if job.UpdatedAt.IsZero() {
				job.UpdatedAt = job.CreatedAt
			}

			if err := inserter.Insert(
				ctx,
				job.BatchSpecWorkspaceID,
				job.CreatedAt,
				job.UpdatedAt,
			); err != nil {
				return err
			}
		}

		return nil
	}
	i := -1
	return batch.WithInserterWithReturn(
		ctx,
		s.Handle().DB(),
		"batch_spec_workspace_execution_jobs",
		[]string{"batch_spec_workspace_id", "created_at", "updated_at"},
		"",
		[]string{
			"batch_spec_workspace_execution_jobs.id",
			"batch_spec_workspace_execution_jobs.batch_spec_workspace_id",
			"batch_spec_workspace_execution_jobs.access_token_id",
			"batch_spec_workspace_execution_jobs.state",
			"batch_spec_workspace_execution_jobs.failure_message",
			"batch_spec_workspace_execution_jobs.started_at",
			"batch_spec_workspace_execution_jobs.finished_at",
			"batch_spec_workspace_execution_jobs.process_after",
			"batch_spec_workspace_execution_jobs.num_resets",
			"batch_spec_workspace_execution_jobs.num_failures",
			"batch_spec_workspace_execution_jobs.execution_logs",
			"batch_spec_workspace_execution_jobs.worker_hostname",
			"batch_spec_workspace_execution_jobs.cancel",
			"NULL as place_in_queue",
			"batch_spec_workspace_execution_jobs.created_at",
			"batch_spec_workspace_execution_jobs.updated_at",
		},
		func(rows *sql.Rows) error {
			i++
			return scanFn(jobs[i], rows)
		},
		inserter,
	)
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
