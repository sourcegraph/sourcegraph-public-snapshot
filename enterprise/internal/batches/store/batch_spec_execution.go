package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var BatchSpecExecutionColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_spec_executions.id"),
	sqlf.Sprintf("batch_spec_executions.state"),
	sqlf.Sprintf("batch_spec_executions.failure_message"),
	sqlf.Sprintf("batch_spec_executions.started_at"),
	sqlf.Sprintf("batch_spec_executions.finished_at"),
	sqlf.Sprintf("batch_spec_executions.process_after"),
	sqlf.Sprintf("batch_spec_executions.num_resets"),
	sqlf.Sprintf("batch_spec_executions.num_failures"),
	sqlf.Sprintf(`batch_spec_executions.execution_logs`),
	sqlf.Sprintf("batch_spec_executions.worker_hostname"),
	sqlf.Sprintf(`batch_spec_executions.created_at`),
	sqlf.Sprintf(`batch_spec_executions.updated_at`),
	sqlf.Sprintf(`batch_spec_executions.batch_spec`),
	sqlf.Sprintf(`batch_spec_executions.batch_spec_id`),
	sqlf.Sprintf(`batch_spec_executions.user_id`),
}

var batchSpecExecutionInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("batch_spec"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

// CreateBatchSpecExecution creates the given BatchSpecExecution.
func (s *Store) CreateBatchSpecExecution(ctx context.Context, b *btypes.BatchSpecExecution) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = s.now()
	}

	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = b.CreatedAt
	}

	q, err := createBatchSpecExecutionQuery(b)
	if err != nil {
		return err
	}
	return s.query(ctx, q, func(sc scanner) error { return scanBatchSpecExecution(b, sc) })
}

var createBatchSpecExecutionQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_executions.go:CreateBatchSpecExecution
INSERT INTO batch_spec_executions (%s)
VALUES (%s, %s, %s, %s)
RETURNING %s`

func createBatchSpecExecutionQuery(c *btypes.BatchSpecExecution) (*sqlf.Query, error) {
	return sqlf.Sprintf(
		createBatchSpecExecutionQueryFmtstr,
		sqlf.Join(batchSpecExecutionInsertColumns, ", "),
		c.BatchSpec,
		c.UserID,
		c.CreatedAt,
		c.UpdatedAt,
		sqlf.Join(BatchSpecExecutionColumns, ", "),
	), nil
}

// GetBatchSpecExecutionOpts captures the query options needed for getting a BatchSpecExecution.
type GetBatchSpecExecutionOpts struct {
	ID int64
}

// GetBatchSpecExecution gets a BatchSpecExecution matching the given options.
func (s *Store) GetBatchSpecExecution(ctx context.Context, opts GetBatchSpecExecutionOpts) (*btypes.BatchSpecExecution, error) {
	q := getBatchSpecExecutionQuery(&opts)

	var b btypes.BatchSpecExecution
	err := s.query(ctx, q, func(sc scanner) (err error) {
		return scanBatchSpecExecution(&b, sc)
	})
	if err != nil {
		return nil, err
	}

	if b.ID == 0 {
		return nil, ErrNoResults
	}

	return &b, nil
}

var getBatchSpecExecutionQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_executions.go:GetBatchSpecExecution
SELECT %s FROM batch_spec_executions
WHERE %s
LIMIT 1
`

func getBatchSpecExecutionQuery(opts *GetBatchSpecExecutionOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBatchSpecExecutionQueryFmtstr,
		sqlf.Join(BatchSpecExecutionColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBatchSpecExecution(b *btypes.BatchSpecExecution, sc scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry

	if err := sc.Scan(
		&b.ID,
		&b.State,
		&b.FailureMessage,
		&b.StartedAt,
		&b.FinishedAt,
		&b.ProcessAfter,
		&b.NumResets,
		&b.NumFailures,
		pq.Array(&executionLogs),
		&b.WorkerHostname,
		&b.CreatedAt,
		&b.UpdatedAt,
		&b.BatchSpec,
		&dbutil.NullInt64{N: &b.BatchSpecID},
		&b.UserID,
	); err != nil {
		return err
	}

	for _, entry := range executionLogs {
		b.ExecutionLogs = append(b.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return nil
}

// scanBatchSpecExecutions scans a slice of batch spec executions from the rows.
func scanBatchSpecExecutions(rows *sql.Rows, queryErr error) (_ []*btypes.BatchSpecExecution, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var execs []*btypes.BatchSpecExecution
	for rows.Next() {
		exec := &btypes.BatchSpecExecution{}
		if err := scanBatchSpecExecution(exec, rows); err != nil {
			return nil, err
		}

		execs = append(execs, exec)
	}

	return execs, nil
}

// ScanFirstBatchSpecExecution scans a slice of batch spec executions from the
// rows and returns the first.
func ScanFirstBatchSpecExecution(rows *sql.Rows, err error) (*btypes.BatchSpecExecution, bool, error) {
	execs, err := scanBatchSpecExecutions(rows, err)
	if err != nil || len(execs) == 0 {
		return &btypes.BatchSpecExecution{}, false, err
	}
	return execs[0], true, nil
}
