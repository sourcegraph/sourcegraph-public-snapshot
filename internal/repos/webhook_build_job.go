package repos

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	basestore "github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	webhookbuilder "github.com/sourcegraph/sourcegraph/internal/repos/worker"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type webhookBuildJob struct {
}

func NewWebhookBuildJob() *webhookBuildJob {
	return &webhookBuildJob{}
}

func (w *webhookBuildJob) Description() string {
	return ""
}

func (w *webhookBuildJob) Config() []env.Config {
	return []env.Config{}
}

func (w *webhookBuildJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("background", "background webhook build job"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	webhookBuildWorkerMetrics, webhookBuildResetterMetrics := newWebhookBuildWorkerMetrics(observationContext, "webhook_build_worker")

	mainAppDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	db := database.NewDB(logger, mainAppDB)
	store := NewStore(logger, db)
	baseStore := basestore.NewWithHandle(store.Handle())
	workerStore := webhookbuilder.CreateWorkerStore(store.Handle())

	return []goroutine.BackgroundRoutine{
		webhookbuilder.NewWorker(ctx, newWebHookBuildHandler(store), workerStore, webhookBuildWorkerMetrics),
		webhookbuilder.NewResetter(ctx, workerStore, webhookBuildResetterMetrics),
		webhookbuilder.NewCleaner(ctx, baseStore, observationContext),
	}, nil
}

func newWebhookBuildWorkerMetrics(observationContext *observation.Context, workerName string) (workerutil.WorkerMetrics, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observationContext, fmt.Sprintf("%s_processor", workerName))
	resetterMetrics := dbworker.NewMetrics(observationContext, workerName)
	return workerMetrics, *resetterMetrics
}
