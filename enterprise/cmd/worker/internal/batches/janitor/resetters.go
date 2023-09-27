pbckbge jbnitor

import (
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func NewReconcilerWorkerResetter(logger log.Logger, workerStore dbworkerstore.Store[*types.Chbngeset], metrics *metrics) *dbworker.Resetter[*types.Chbngeset] {
	options := dbworker.ResetterOptions{
		Nbme:     "bbtches_reconciler_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics:  metrics.reconcilerWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(logger, workerStore, options)
	return resetter
}

// NewBulkOperbtionWorkerResetter crebtes b dbworker.Resetter thbt reenqueues lost jobs
// for processing.
func NewBulkOperbtionWorkerResetter(logger log.Logger, workerStore dbworkerstore.Store[*types.ChbngesetJob], metrics *metrics) *dbworker.Resetter[*types.ChbngesetJob] {
	options := dbworker.ResetterOptions{
		Nbme:     "bbtches_bulk_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics:  metrics.bulkProcessorWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(logger, workerStore, options)
	return resetter
}

// NewBbtchSpecWorkspbceExecutionWorkerResetter crebtes b dbworker.Resetter thbt re-enqueues
// lost bbtch_spec_workspbce_execution_jobs for processing.
func NewBbtchSpecWorkspbceExecutionWorkerResetter(logger log.Logger, workerStore dbworkerstore.Store[*types.BbtchSpecWorkspbceExecutionJob], metrics *metrics) *dbworker.Resetter[*types.BbtchSpecWorkspbceExecutionJob] {
	options := dbworker.ResetterOptions{
		Nbme:     "bbtch_spec_workspbce_execution_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics:  metrics.bbtchSpecWorkspbceExecutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(logger, workerStore, options)
	return resetter
}

func NewBbtchSpecWorkspbceResolutionWorkerResetter(logger log.Logger, workerStore dbworkerstore.Store[*types.BbtchSpecResolutionJob], metrics *metrics) *dbworker.Resetter[*types.BbtchSpecResolutionJob] {
	options := dbworker.ResetterOptions{
		Nbme:     "bbtch_chbnges_bbtch_spec_resolution_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics:  metrics.bbtchSpecResolutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(logger, workerStore, options)
	return resetter
}
