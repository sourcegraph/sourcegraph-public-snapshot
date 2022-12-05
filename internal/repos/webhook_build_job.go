package repos

import (
	"context"
	"fmt"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// webhookBuildJob implements the Job interface
// from package job
type webhookBuildJob struct{}

func NewWebhookBuildJob() *webhookBuildJob {
	return &webhookBuildJob{}
}

func (w *webhookBuildJob) Description() string {
	return "A background routine that builds webhooks for repos"
}

func (w *webhookBuildJob) Config() []env.Config {
	return []env.Config{}
}

func (w *webhookBuildJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	webhookBuildWorkerMetrics, webhookBuildResetterMetrics := newWebhookBuildWorkerMetrics(observationCtx, "webhook_build_worker")

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	store := NewStore(observationCtx.Logger, db)
	baseStore := basestore.NewWithHandle(store.Handle())
	workerStore := webhookworker.CreateWorkerStore(observationCtx, store.Handle())

	cf := httpcli.NewExternalClientFactory(httpcli.NewLoggingMiddleware(observationCtx.Logger))
	doer, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}

	return []goroutine.BackgroundRoutine{
		webhookworker.NewWorker(context.Background(), newWebhookBuildHandler(store, doer), workerStore, webhookBuildWorkerMetrics),
		webhookworker.NewResetter(context.Background(), observationCtx.Logger, workerStore, webhookBuildResetterMetrics),
		webhookworker.NewCleaner(context.Background(), observationCtx, baseStore),
	}, nil
}

func newWebhookBuildWorkerMetrics(observationCtx *observation.Context, workerName string) (workerutil.WorkerObservability, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observationCtx, fmt.Sprintf("%s_processor", workerName))
	resetterMetrics := dbworker.NewMetrics(observationCtx, workerName)
	return workerMetrics, *resetterMetrics
}
