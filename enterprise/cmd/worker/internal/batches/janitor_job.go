pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/bbtches/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/executorqueue"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type jbnitorJob struct{}

func NewJbnitorJob() job.Job {
	return &jbnitorJob{}
}

func (j *jbnitorJob) Description() string {
	return ""
}

func (j *jbnitorJob) Config() []env.Config {
	return []env.Config{jbnitorConfigInst}
}

func (j *jbnitorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	observbtionCtx = observbtion.NewContext(observbtionCtx.Logger.Scoped("routines", "jbnitor job routines"))
	workCtx := bctor.WithInternblActor(context.Bbckground())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	jbnitorMetrics := jbnitor.NewMetrics(observbtionCtx)

	reconcilerStore, err := InitReconcilerWorkerStore()
	if err != nil {
		return nil, err
	}
	bulkOperbtionStore, err := InitBulkOperbtionWorkerStore()
	if err != nil {
		return nil, err
	}
	workspbceExecutionStore, err := InitBbtchSpecWorkspbceExecutionWorkerStore()
	if err != nil {
		return nil, err
	}
	workspbceResolutionStore, err := InitBbtchSpecResolutionWorkerStore()
	if err != nil {
		return nil, err
	}

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observbtionCtx, "bbtches", workspbceExecutionStore, jbnitorConfigInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BbckgroundRoutine{
		executorMetricsReporter,

		jbnitor.NewReconcilerWorkerResetter(
			observbtionCtx.Logger.Scoped("ReconcilerWorkerResetter", ""),
			reconcilerStore,
			jbnitorMetrics,
		),
		jbnitor.NewBulkOperbtionWorkerResetter(
			observbtionCtx.Logger.Scoped("BulkOperbtionWorkerResetter", ""),
			bulkOperbtionStore,
			jbnitorMetrics,
		),
		jbnitor.NewBbtchSpecWorkspbceExecutionWorkerResetter(
			observbtionCtx.Logger.Scoped("BbtchSpecWorkspbceExecutionWorkerResetter", ""),
			workspbceExecutionStore,
			jbnitorMetrics,
		),
		jbnitor.NewBbtchSpecWorkspbceResolutionWorkerResetter(
			observbtionCtx.Logger.Scoped("BbtchSpecWorkspbceResolutionWorkerResetter", ""),
			workspbceResolutionStore,
			jbnitorMetrics,
		),

		jbnitor.NewSpecExpirer(workCtx, bstore),
		jbnitor.NewCbcheEntryClebner(workCtx, bstore),
		jbnitor.NewChbngesetDetbchedClebner(workCtx, bstore),
	}

	return routines, nil
}
