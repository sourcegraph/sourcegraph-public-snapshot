package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const batchSpecWorkspaceExecutionJobStalledJobMaximumAge = time.Second * 25
const batchSpecWorkspaceExecutionJobMaximumNumResets = 3

func scanFirstBatchSpecWorkspaceExecutionJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecWorkspaceExecutionJob(rows, err)
}

// newBatchSpecWorkspaceExecutionWorkerResetter creates a dbworker.Resetter that re-enqueues
// lost batch_spec_workspace_execution_jobs for processing.
func newBatchSpecWorkspaceExecutionWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_spec_workspace_execution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecWorkspaceExecutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

var batchSpecWorkspaceExecutionWorkerStoreOptions = dbworkerstore.Options{
	Name:              "batch_spec_workspace_execution_worker_store",
	TableName:         "batch_spec_workspace_execution_jobs",
	ColumnExpressions: store.BatchSpecWorkspaceExecutionJobColums.ToSqlf(),
	Scan:              scanFirstBatchSpecWorkspaceExecutionJobRecord,
	OrderByExpression: sqlf.Sprintf("batch_spec_workspace_execution_jobs.created_at, batch_spec_workspace_execution_jobs.id"),
	StalledMaxAge:     batchSpecWorkspaceExecutionJobStalledJobMaximumAge,
	MaxNumResets:      batchSpecWorkspaceExecutionJobMaximumNumResets,
	// Explicitly disable retries.
	MaxNumRetries: 0,
}

type BatchSpecWorkspaceExecutionWorkerStore interface {
	dbworkerstore.Store
	FetchCanceled(ctx context.Context, executorName string) (canceledIDs []int, err error)
}

// NewBatchSpecWorkspaceExecutionWorkerStore creates a dbworker store that
// wraps the batch_spec_workspace_execution_jobs table.
func NewBatchSpecWorkspaceExecutionWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) BatchSpecWorkspaceExecutionWorkerStore {
	return &batchSpecWorkspaceExecutionWorkerStore{
		Store:              dbworkerstore.NewWithMetrics(handle, batchSpecWorkspaceExecutionWorkerStoreOptions, observationContext),
		observationContext: observationContext,
	}
}

var _ dbworkerstore.Store = &batchSpecWorkspaceExecutionWorkerStore{}

// batchSpecWorkspaceExecutionWorkerStore is a thin wrapper around
// dbworkerstore.Store that allows us to extract information out of the
// ExecutionLogEntry field and persisting it to separate columns when marking a
// job as complete.
type batchSpecWorkspaceExecutionWorkerStore struct {
	dbworkerstore.Store

	observationContext *observation.Context
}

func (s *batchSpecWorkspaceExecutionWorkerStore) FetchCanceled(ctx context.Context, executorName string) (canceledIDs []int, err error) {
	batchesStore := store.New(s.Store.Handle().DB(), s.observationContext, nil)

	t := true
	cs, err := batchesStore.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{
		Cancel:         &t,
		State:          btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		WorkerHostname: executorName,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(cs))
	for _, c := range cs {
		ids = append(ids, c.RecordID())
	}
	return ids, nil
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkComplete(ctx context.Context, id int, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	batchesStore := store.New(s.Store.Handle().DB(), s.observationContext, nil)

	tx, err := batchesStore.Transact(ctx)
	if err != nil {
		return false, err
	}
	defer func() { err = tx.Done(err) }()

	job, changesetSpecIDs, err := loadAndExtractChangesetSpecIDs(ctx, tx, int64(id))
	if err != nil {
		// If we couldn't extract the changeset IDs, we mark the job as failed
		return s.Store.MarkFailed(ctx, id, fmt.Sprintf("failed to extract changeset IDs ID: %s", err), options)
	}

	return markBatchSpecWorkspaceExecutionJobComplete(ctx, tx, job, changesetSpecIDs, options.WorkerHostname)
}

// markBatchSpecWorkspaceExecutionJobCompleteQuery is taken from internal/workerutil/dbworker/store/store.go
//
// If that one changes we need to update this one here too.
const markBatchSpecWorkspaceExecutionJobCompleteQuery = `
UPDATE batch_spec_workspace_execution_jobs
SET state = 'completed', finished_at = clock_timestamp()
WHERE id = %s AND state = 'processing' AND worker_hostname = %s
RETURNING id
`

const setChangesetSpecIDsOnBatchSpecWorkspace = `
UPDATE batch_spec_workspaces SET changeset_spec_ids = %s WHERE id = %s
`

const setBatchSpecIDOnChangesetSpecs = `
UPDATE changeset_specs
SET batch_spec_id = (SELECT batch_spec_id FROM batch_spec_workspaces WHERE id = %s LIMIT 1)
WHERE id IN (%s)
`

func markBatchSpecWorkspaceExecutionJobComplete(ctx context.Context, tx *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob, changesetSpecIDs []int64, workerHostname string) (bool, error) {
	ids := []*sqlf.Query{}
	m := make(map[int64]struct{}, len(changesetSpecIDs))
	for _, id := range changesetSpecIDs {
		ids = append(ids, sqlf.Sprintf("%s", id))
		m[id] = struct{}{}
	}

	// Set the batch_spec_id on the changeset_specs that were created
	err := tx.Exec(ctx, sqlf.Sprintf(setBatchSpecIDOnChangesetSpecs, job.BatchSpecWorkspaceID, sqlf.Join(ids, ",")))
	if err != nil {
		return false, err
	}

	marshaledIDs, err := json.Marshal(m)
	if err != nil {
		return false, err
	}

	// Set changeset_spec_ids on the batch_spec_workspace
	err = tx.Exec(ctx, sqlf.Sprintf(setChangesetSpecIDsOnBatchSpecWorkspace, marshaledIDs, job.BatchSpecWorkspaceID))
	if err != nil {
		return false, err
	}

	// Finally mark batch_spec_workspace_execution_jobs as completed
	_, ok, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(markBatchSpecWorkspaceExecutionJobCompleteQuery, job.ID, workerHostname)))
	return ok, err
}

func loadAndExtractChangesetSpecIDs(ctx context.Context, s *store.Store, id int64) (*btypes.BatchSpecWorkspaceExecutionJob, []int64, error) {
	job, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, store.GetBatchSpecWorkspaceExecutionJobOpts{ID: id})
	if err != nil {
		return job, []int64{}, err
	}

	if len(job.ExecutionLogs) < 1 {
		return job, []int64{}, errors.Newf("job %d has no execution logs", job.ID)
	}

	randIDs, err := extractChangesetSpecRandIDs(job.ExecutionLogs)
	if err != nil {
		return job, []int64{}, err
	}

	specs, _, err := s.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{LimitOpts: store.LimitOpts{Limit: 0}, RandIDs: randIDs})
	if err != nil {
		return job, []int64{}, err
	}

	var ids []int64
	for _, spec := range specs {
		ids = append(ids, spec.ID)
	}

	return job, ids, nil
}

var ErrNoChangesetSpecIDs = errors.New("no changeset ids found in execution logs")

func extractChangesetSpecRandIDs(logs []workerutil.ExecutionLogEntry) ([]string, error) {
	var (
		randIDs []string
		entry   workerutil.ExecutionLogEntry
		found   bool
	)

	for _, e := range logs {
		if e.Key == "step.src.0" {
			entry = e
			found = true
			break
		}
	}
	if !found {
		return randIDs, ErrNoChangesetSpecIDs
	}

	for _, l := range strings.Split(entry.Out, "\n") {
		const outputLinePrefix = "stdout: "

		if !strings.HasPrefix(l, outputLinePrefix) {
			continue
		}

		jsonPart := l[len(outputLinePrefix):]

		var e changesetSpecsUploadedLogLine
		if err := json.Unmarshal([]byte(jsonPart), &e); err != nil {
			// If we can't unmarshal the line as JSON we skip it
			continue
		}

		if e.Operation == operationUploadingChangesetSpecs && e.Status == "SUCCESS" {
			rawIDs := e.Metadata.IDs
			if len(rawIDs) == 0 {
				return randIDs, ErrNoChangesetSpecIDs
			}

			var randIDs []string
			for _, rawID := range rawIDs {
				var randID string
				if err := relay.UnmarshalSpec(graphql.ID(rawID), &randID); err != nil {
					return randIDs, errors.Wrap(err, "failed to unmarshal changeset spec rand id")
				}

				randIDs = append(randIDs, randID)
			}

			return randIDs, nil
		}
	}

	return randIDs, ErrNoChangesetSpecIDs
}

type changesetSpecsUploadedLogLine struct {
	Operation string
	Timestamp time.Time
	Status    string
	Metadata  struct {
		IDs []string `json:"ids"`
	}
}

const operationUploadingChangesetSpecs = "UPLOADING_CHANGESET_SPECS"
