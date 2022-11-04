package repos

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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

func (w *webhookBuildJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("background", "background webhook build job"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	webhookBuildWorkerMetrics, webhookBuildResetterMetrics := newWebhookBuildWorkerMetrics(observationContext, "webhook_build_worker")

	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return nil, err
	}

	store := NewStore(logger, db)
	baseStore := basestore.NewWithHandle(store.Handle())
	workerStore := webhookworker.CreateWorkerStore(logger.Scoped("webhookworker.WorkerStore", ""), store.Handle())

	cf := httpcli.NewExternalClientFactory(httpcli.NewLoggingMiddleware(logger))
	doer, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}

	return []goroutine.BackgroundRoutine{
		webhookworker.NewWorker(context.Background(), newWebhookBuildHandler(store, doer), workerStore, webhookBuildWorkerMetrics),
		webhookworker.NewResetter(context.Background(), logger.Scoped("webhookworker.Resetter", ""), workerStore, webhookBuildResetterMetrics),
		webhookworker.NewCleaner(context.Background(), baseStore, observationContext),
	}, nil
}

func newWebhookBuildWorkerMetrics(observationContext *observation.Context, workerName string) (workerutil.WorkerObservability, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observationContext, fmt.Sprintf("%s_processor", workerName))
	resetterMetrics := dbworker.NewMetrics(observationContext, workerName)
	return workerMetrics, *resetterMetrics
}
