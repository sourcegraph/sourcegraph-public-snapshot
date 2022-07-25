package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"

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

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), time.Minute*1, newTelemetryHandler(logger, eventLogStore, sendEvents), operation)
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

const MaxEventsCountDefault = 5

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

func sendEvents(ctx context.Context, events []*types.Event) error {
	client, err := pubsub.NewClient(ctx, "sourcegraph-dogfood")
	if err != nil {
		log15.Error("failed to create pubsub client.", "error", err)
		return err
	}

	var toSend []bigQueryEvent
	for _, event := range events {
		pubsubEvent := bigQueryEvent{
			EventName:       event.Name,
			UserID:          int(*event.UserID),
			AnonymousUserID: event.AnonymousUserID,
			URL:             event.URL,
			Source:          event.Source,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			Version:         version.Version(),
			PublicArgument:  event.Argument,
		}
		toSend = append(toSend, pubsubEvent)
	}

	msg, err := json.Marshal(toSend)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	topic := client.Topic("usage-data-testing")
	defer topic.Stop()
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})
	_, err = result.Get(ctx)
	if err != nil {
		return err
	}

	log15.Info("sent pubsub message")
	return nil
}

type bigQueryEvent struct {
	EventName       string  `json:"name"`
	URL             string  `json:"url"`
	AnonymousUserID string  `json:"anonymous_user_id"`
	FirstSourceURL  string  `json:"first_source_url"`
	LastSourceURL   string  `json:"last_source_url"`
	UserID          int     `json:"user_id"`
	Source          string  `json:"source"`
	Timestamp       string  `json:"timestamp"`
	Version         string  `json:"version"`
	FeatureFlags    string  `json:"feature_flags"`
	CohortID        *string `json:"cohort_id,omitempty"`
	Referrer        string  `json:"referrer,omitempty"`
	PublicArgument  string  `json:"public_argument"`
	DeviceID        *string `json:"device_id,omitempty"`
	InsertID        *string `json:"insert_id,omitempty"`
}

var pubSubDotComEventsTopicID = env.Get("USAGE_DATA_EVENTS_TOPIC_ID", "", "Pub/sub events topic ID is the pub/sub topic id where usage data events are published.")
