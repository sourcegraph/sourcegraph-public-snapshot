package telemetry

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type telemetryJob struct{}

func NewTelemetryJob() *telemetryJob {
	return &telemetryJob{}
}

func (t *telemetryJob) Description() string {
	return "A background routine that exports usage telemetry to Sourcegraph"
}

func (t *telemetryJob) Config() []env.Config {
	return nil
}

func (t *telemetryJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	if !isEnabled() {
		return nil, nil
	}
	logger.Info("Usage telemetry export enabled - initializing background routine")

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		newBackgroundTelemetryJob(database.NewDB(logger, db), logger),
	}, nil
}

func newBackgroundTelemetryJob(db database.DB, logger log.Logger) goroutine.BackgroundRoutine {
	observationContext := &observation.Context{
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.NewRegistry(),
	}
	operation := observationContext.Operation(observation.Op{})

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), time.Minute*1, &telemetryHandler{db: db, logger: logger}, operation)
}

type telemetryHandler struct {
	db     database.DB
	logger log.Logger
}

var disabledErr = errors.New("Usage telemetry export is disabled, but the background job is attempting to execute. This means the configuration was disabled without restarting the worker service. This job is aborting, and no telemetry will be exported.")

func (t *telemetryHandler) Handle(ctx context.Context) error {
	if !isEnabled() {
		return disabledErr
	}

	t.logger.Info("telemetryHandler executed")
	return nil
}

func isEnabled() bool {
	ptr := conf.Get().ExportUsageTelemetry
	if ptr != nil {
		return ptr.Enabled
	}

	return false
}
