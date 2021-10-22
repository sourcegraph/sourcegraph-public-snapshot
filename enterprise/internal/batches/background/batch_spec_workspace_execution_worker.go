package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// batchSpecWorkspaceExecutionJobStalledJobMaximumAge is the maximum allowable
// duration between updating the state of a job as "processing" and locking the
// record during processing. An unlocked row that is marked as processing
// likely indicates that the executor that dequeued the job has died. There
// should be a nearly-zero delay between these states during normal operation.
const batchSpecWorkspaceExecutionJobStalledJobMaximumAge = time.Second * 25

// batchSpecWorkspaceExecutionJobMaximumNumResets is the maximum number of
// times a job can be reset. If a job's failed attempts counter reaches this
// threshold, it will be moved into "errored" rather than "queued" on its next
// reset.
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
	// This needs to be kept in sync with the placeInQueue fragment in the batch
	// spec execution jobs store.
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

// deleteAccessToken tries to delete the associated internal access
// token. If the token cannot be found it does *not* return an error.
func deleteAccessToken(ctx context.Context, batchesStore *store.Store, tokenID int64) error {
	err := database.AccessTokensWith(batchesStore).HardDeleteByID(ctx, tokenID)
	if err != nil && err != database.ErrAccessTokenNotFound {
		return err
	}
	return nil
}

type markFinal func(ctx context.Context, tx dbworkerstore.Store) (_ bool, err error)

func (s *batchSpecWorkspaceExecutionWorkerStore) deleteAccessTokenAndMarkFinal(ctx context.Context, id int, failureMessage string, options dbworkerstore.MarkFinalOptions, fn markFinal) (_ bool, err error) {
	batchesStore := store.New(s.Store.Handle().DB(), s.observationContext, nil)
	tx, err := batchesStore.Transact(ctx)
	if err != nil {
		return false, err
	}
	defer func() { err = tx.Done(err) }()

	job, err := batchesStore.GetBatchSpecWorkspaceExecutionJob(ctx, store.GetBatchSpecWorkspaceExecutionJobOpts{ID: int64(id)})
	if err != nil {
		return false, err
	}
	err = deleteAccessToken(ctx, batchesStore, job.AccessTokenID)
	if err != nil {
		return false, err
	}
	return fn(ctx, s.Store.With(tx))
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkErrored(ctx context.Context, id int, failureMessage string, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	return s.deleteAccessTokenAndMarkFinal(ctx, id, failureMessage, options, func(ctx context.Context, tx dbworkerstore.Store) (bool, error) {
		return tx.MarkErrored(ctx, id, failureMessage, options)
	})
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkFailed(ctx context.Context, id int, failureMessage string, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	return s.deleteAccessTokenAndMarkFinal(ctx, id, failureMessage, options, func(ctx context.Context, tx dbworkerstore.Store) (bool, error) {
		return tx.MarkFailed(ctx, id, failureMessage, options)
	})
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkComplete(ctx context.Context, id int, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	batchesStore := store.New(s.Store.Handle().DB(), s.observationContext, nil)

	tx, err := batchesStore.Transact(ctx)
	if err != nil {
		return false, err
	}

	job, changesetSpecIDs, err := loadAndExtractChangesetSpecIDs(ctx, tx, int64(id))
	if err != nil {
		// Rollback transaction but ignore rollback errors
		tx.Done(err)
		return s.Store.With(tx).MarkFailed(ctx, id, fmt.Sprintf("failed to extract changeset IDs ID: %s", err), options)
	}

	err = deleteAccessToken(ctx, tx, job.AccessTokenID)
	if err != nil {
		// Rollback transaction but ignore rollback errors
		tx.Done(err)
		// If we failed do delete the access token, we don't need to try again
		// in our MarkFailed method.
		return s.Store.MarkFailed(ctx, id, fmt.Sprintf("failed to delete internal access token: %s", err), options)
	}

	err = setChangesetSpecIDs(ctx, tx, job.BatchSpecWorkspaceID, changesetSpecIDs)
	if err != nil {
		return false, tx.Done(err)
	}

	ok, err := s.Store.With(tx).MarkComplete(ctx, id, options)
	return ok, tx.Done(err)
}

const setChangesetSpecIDsOnBatchSpecWorkspace = `
UPDATE batch_spec_workspaces SET changeset_spec_ids = %s WHERE id = %s
`

const setBatchSpecIDOnChangesetSpecs = `
UPDATE changeset_specs
SET batch_spec_id = (SELECT batch_spec_id FROM batch_spec_workspaces WHERE id = %s LIMIT 1)
WHERE id = ANY (%s)
`

func setChangesetSpecIDs(ctx context.Context, tx *store.Store, batchSpecWorkspaceID int64, changesetSpecIDs []int64) error {
	// Set the batch_spec_id on the changeset_specs that were created
	err := tx.Exec(ctx, sqlf.Sprintf(setBatchSpecIDOnChangesetSpecs, batchSpecWorkspaceID, pq.Array(changesetSpecIDs)))
	if err != nil {
		return err
	}

	m := make(map[int64]struct{}, len(changesetSpecIDs))
	for _, id := range changesetSpecIDs {
		m[id] = struct{}{}
	}
	marshaledIDs, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Set changeset_spec_ids on the batch_spec_workspace
	return tx.Exec(ctx, sqlf.Sprintf(setChangesetSpecIDsOnBatchSpecWorkspace, marshaledIDs, batchSpecWorkspaceID))
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
		entry workerutil.ExecutionLogEntry
		found bool
	)

	for _, e := range logs {
		if e.Key == "step.src.0" {
			entry = e
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNoChangesetSpecIDs
	}

	logLines := btypes.ParseJSONLogsFromOutput(entry.Out)
	for _, l := range logLines {
		if l.Status != batcheslib.LogEventStatusSuccess {
			continue
		}
		if m, ok := l.Metadata.(*batcheslib.UploadingChangesetSpecsMetadata); ok {
			rawIDs := m.IDs
			if len(rawIDs) == 0 {
				return nil, ErrNoChangesetSpecIDs
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

	return nil, ErrNoChangesetSpecIDs
}
