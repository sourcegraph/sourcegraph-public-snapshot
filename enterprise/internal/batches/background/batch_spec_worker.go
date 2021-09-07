package background

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

const batchspecMaxNumRetries = 60
const batchspecMaxNumResets = 60

// newBatchSpecWorker creates a dbworker.newWorker that fetches enqueued changesets
// from the database and passes them to the changeset batchspec for
// processing.
func newBatchSpecWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	e := &evaluator{store: s}

	options := workerutil.WorkerOptions{
		Name:              "batches_batchspec_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.reconcilerWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, e.HandlerFunc(), options)
	return worker
}

func newBatchSpecWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_batch_spec_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func scanFirstBatchSpecRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpec(rows, err)
}

func NewBatchSpecDBWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	options := dbworkerstore.Options{
		Name:              "batches_batch_spec_worker_store",
		TableName:         "batch_specs",
		ColumnExpressions: store.BatchSpecColumns,
		Scan:              scanFirstBatchSpecRecord,

		OrderByExpression: sqlf.Sprintf("batch_specs.state = 'errored', batch_specs.updated_at DESC"),

		StalledMaxAge: 60 * time.Second,
		MaxNumResets:  batchspecMaxNumResets,

		RetryAfter:    5 * time.Second,
		MaxNumRetries: batchspecMaxNumRetries,
	}

	return dbworkerstore.NewWithMetrics(handle, options, observationContext)
}

type evaluator struct {
	store *store.Store
}

// HandlerFunc returns a dbworker.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (e *evaluator) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, record workerutil.Record) (err error) {
		tx, err := e.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() {
			doneErr := tx.Done(nil)
			if err != nil && doneErr != nil {
				err = multierror.Append(err, doneErr)
			}
			if doneErr != nil {
				err = doneErr
			}
		}()

		return e.process(ctx, tx, record.(*btypes.BatchSpec))
	}
}

func (r *evaluator) process(ctx context.Context, tx *store.Store, spec *btypes.BatchSpec) error {
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("---- PROCESSING BATCH SPEC %d ----\n", spec.ID)

	evaluatableSpec, err := batcheslib.ParseBatchSpec([]byte(spec.RawSpec), batcheslib.ParseBatchSpecOptions{
		AllowArrayEnvironments: true,
		AllowTransformChanges:  true,
		AllowConditionalExec:   true,
	})
	if err != nil {
		return err
	}

	workspaces, unsupported, ignored, err := service.New(tx).ResolveWorkspacesForBatchSpec(ctx, evaluatableSpec, service.ResolveWorkspacesForBatchSpecOpts{
		// TODO: Do we need to persist those on the batch spec?
		AllowIgnored:     true,
		AllowUnsupported: true,
	})
	if err != nil {
		return err
	}

	fmt.Printf("----  len(workspaces)=%d, len(unsupported)=%d, len(ignored)=%d \n", len(workspaces), len(unsupported), len(ignored))

	var workspaceJobs []*btypes.BatchSpecWorkspaceJob
	for _, w := range workspaces {
		workspaceJobs = append(workspaceJobs, &btypes.BatchSpecWorkspaceJob{
			BatchSpecID:      spec.ID,
			ChangesetSpecIDs: []int64{},
			RepoID:           w.Repo.ID,
			Branch:           w.Branch,
			Commit:           string(w.Commit),
			Path:             w.Path,
			State:            btypes.BatchSpecWorkspaceJobStatePending,
		})
	}

	return tx.CreateBatchSpecWorkspaceJob(ctx, workspaceJobs...)
}
