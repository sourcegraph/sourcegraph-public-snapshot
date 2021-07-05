package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// executorStalledJobMaximumAge is the maximum allowable duration between updating the state of a
// job as "processing" and locking the record during processing. An unlocked row that is
// marked as processing likely indicates that the executor that dequeued the job has died.
// There should be a nearly-zero delay between these states during normal operation.
const executorStalledJobMaximumAge = time.Second * 5

// executorMaximumNumResets is the maximum number of times a job can be reset. If a job's failed
// attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const executorMaximumNumResets = 3

var executorWorkerStoreOptions = dbworkerstore.Options{
	Name:              "batch_spec_executor_worker_store",
	TableName:         "batch_spec_executions",
	ColumnExpressions: store.BatchSpecExecutionColumns,
	Scan:              scanFirstExecutionRecord,
	OrderByExpression: sqlf.Sprintf("batch_spec_executions.created_at, batch_spec_executions.id"),
	StalledMaxAge:     executorStalledJobMaximumAge,
	MaxNumResets:      executorMaximumNumResets,
	// Explicitly disable retries.
	MaxNumRetries: 0,
}

// NewExecutorStore creates a dbworker store that wraps the batch_spec_executions
// table.
func NewExecutorStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return &executorStore{wrapped: dbworkerstore.NewWithMetrics(s.Handle(), executorWorkerStoreOptions, observationContext)}
}

var _ dbworkerstore.Store = &executorStore{}

type executorStore struct {
	wrapped dbworkerstore.Store
}

const markCompleteQuery = `
UPDATE batch_spec_executions
SET state = 'completed', finished_at = clock_timestamp(), batch_spec_id = (SELECT id FROM batch_specs WHERE rand_id = %s)
WHERE id = %s AND state = 'processing'
RETURNING id
`

func (s *executorStore) MarkComplete(ctx context.Context, id int) (_ bool, err error) {
	batchesStore := store.New(s.wrapped.Handle().DB(), nil)

	exec, err := batchesStore.GetBatchSpecExecution(ctx, store.GetBatchSpecExecutionOpts{ID: int64(id)})

	lines := strings.Split(strings.ReplaceAll(exec.ExecutionLogs[0].Out, "stdout: ", ""), "\n")

	type batchesLogEvent struct {
		Operation string `json:"operation"` // "PREPARING_DOCKER_IMAGES"

		Timestamp time.Time `json:"timestamp"`

		Status  string `json:"status"`            // "STARTED", "PROGRESS", "SUCCESS", "FAILURE"
		Message string `json:"message,omitempty"` // "70% done"
	}

	var batchSpecRandID string
	for _, l := range lines {
		var e batchesLogEvent

		if err := json.Unmarshal([]byte(l), &e); err != nil {
			return false, err
		}

		if e.Operation == "CREATING_BATCH_SPEC" && e.Status == "SUCCESS" {
			parts := strings.Split(e.Message, "/")
			batchSpecGraphQLID := graphql.ID(parts[len(parts)-1])

			if err := relay.UnmarshalSpec(batchSpecGraphQLID, &batchSpecRandID); err != nil {
				return false, err
			}
			break
		}
	}

	h := basestore.NewWithHandle(s.wrapped.Handle())
	_, ok, err := basestore.ScanFirstInt(h.Query(ctx, sqlf.Sprintf(markCompleteQuery, batchSpecRandID, id)))

	return ok, err
}

func (s *executorStore) DequeueWithIndependentTransactionContext(ctx context.Context, workerHostname string, conditions []*sqlf.Query) (workerutil.Record, dbworkerstore.Store, bool, error) {
	r, wrapped, b, err := s.wrapped.DequeueWithIndependentTransactionContext(ctx, workerHostname, conditions)

	return r, &executorStore{wrapped: wrapped}, b, err
}

func (s *executorStore) Dequeue(ctx context.Context, workerHostname string, conditions []*sqlf.Query) (record workerutil.Record, tx dbworkerstore.Store, exists bool, err error) {
	r, wrapped, b, err := s.wrapped.Dequeue(ctx, workerHostname, conditions)

	return r, &executorStore{wrapped: wrapped}, b, err
}
func (s *executorStore) Handle() *basestore.TransactableHandle {
	return s.wrapped.Handle()
}
func (s *executorStore) Done(err error) error {
	return s.wrapped.Done(err)
}
func (s *executorStore) QueuedCount(ctx context.Context, conditions []*sqlf.Query) (int, error) {
	return s.wrapped.QueuedCount(ctx, conditions)
}
func (s *executorStore) Requeue(ctx context.Context, id int, after time.Time) error {
	return s.wrapped.Requeue(ctx, id, after)
}
func (s *executorStore) AddExecutionLogEntry(ctx context.Context, id int, entry workerutil.ExecutionLogEntry) error {
	return s.wrapped.AddExecutionLogEntry(ctx, id, entry)
}
func (s *executorStore) MarkErrored(ctx context.Context, id int, failureMessage string) (bool, error) {
	return s.wrapped.MarkErrored(ctx, id, failureMessage)
}
func (s *executorStore) MarkFailed(ctx context.Context, id int, failureMessage string) (bool, error) {
	return s.wrapped.MarkFailed(ctx, id, failureMessage)
}
func (s *executorStore) ResetStalled(ctx context.Context) (resetIDs, erroredIDs []int, err error) {
	return s.wrapped.ResetStalled(ctx)
}

// scanFirstExecutionRecord scans a slice of batch change executions and returns the first.
func scanFirstExecutionRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecExecution(rows, err)
}
