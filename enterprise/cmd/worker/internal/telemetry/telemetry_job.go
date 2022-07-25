package telemetry

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type telemetryJob struct {
}

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

	sqlDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	db := database.NewDB(logger, sqlDB)
	eventLogStore := db.EventLogs()

	return []goroutine.BackgroundRoutine{
		newBackgroundTelemetryJob(logger, eventLogStore),
	}, nil
}

func newBackgroundTelemetryJob(logger log.Logger, eventLogStore database.EventLogStore) goroutine.BackgroundRoutine {
	observationContext := &observation.Context{
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.NewRegistry(),
	}
	operation := observationContext.Operation(observation.Op{})

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), time.Minute*1, newTelemetryHandler(logger, eventLogStore, func(ctx context.Context, event []*types.Event) error {
		return nil
	}), operation)
}

type telemetryHandler struct {
	logger             log.Logger
	eventLogStore      database.EventLogStore
	sendEventsCallback func(ctx context.Context, event []*types.Event) error
}

func newTelemetryHandler(logger log.Logger, store database.EventLogStore, sendEventsCallback func(ctx context.Context, event []*types.Event) error) *telemetryHandler {
	return &telemetryHandler{
		logger:             logger,
		eventLogStore:      store,
		sendEventsCallback: sendEventsCallback,
	}
}

var disabledErr = errors.New("Usage telemetry export is disabled, but the background job is attempting to execute. This means the configuration was disabled without restarting the worker service. This job is aborting, and no telemetry will be exported.")

const MaxEventsCountDefault = 5000

func (t *telemetryHandler) Handle(ctx context.Context) error {
	if !isEnabled() {
		return disabledErr
	}

	batchSize := getBatchSize()
	all, err := t.eventLogStore.ListExportableEvents(ctx, database.LimitOffset{
		Limit:  batchSize,
		Offset: 0, // currently static, will become dynamic with https://github.com/sourcegraph/sourcegraph/issues/39089
	})
	if err != nil {
		return errors.Wrap(err, "eventLogStore.ListExportableEvents")
	}
	if len(all) == 0 {
		return nil
	}

	maxId := int(all[len(all)-1].ID)
	t.logger.Info("telemetryHandler executed", log.Int("event count", len(all)), log.Int("maxId", maxId))
	return t.sendEventsCallback(ctx, all)
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

func getBatchSize() int {
	val := confClient.Get().ExportUsageTelemetry.BatchSize
	if val <= 0 {
		val = MaxEventsCountDefault
	}
	return val
}
