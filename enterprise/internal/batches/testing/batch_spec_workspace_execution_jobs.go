package testing

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
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

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
