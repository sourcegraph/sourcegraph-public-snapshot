package telemetry

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
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

	return []goroutine.BackgroundRoutine{
		newBackgroundTelemetryJob(logger),
	}, nil
}

func newBackgroundTelemetryJob(logger log.Logger) goroutine.BackgroundRoutine {
	observationContext := &observation.Context{
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.NewRegistry(),
	}
	operation := observationContext.Operation(observation.Op{})

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), time.Minute*1, &telemetryHandler{logger: logger}, operation)
}

type telemetryHandler struct {
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

// This package level client is to prevent race conditions when mocking this configuration in tests.
var confClient = conf.DefaultClient()

func isEnabled() bool {
	ptr := confClient.Get().ExportUsageTelemetry
	if ptr != nil {
		return ptr.Enabled
	}

	return false
}
