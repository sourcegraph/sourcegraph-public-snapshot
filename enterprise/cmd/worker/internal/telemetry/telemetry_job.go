package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/metrics"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"

	"github.com/sourcegraph/sourcegraph/internal/version"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func (t *telemetryJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !isEnabled() {
		return nil, nil
	}
	observationCtx.Logger.Info("Usage telemetry export enabled - initializing background routine")

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		newBackgroundTelemetryJob(observationCtx.Logger, db),
		queueSizeMetricJob(db),
	}, nil
}

func queueSizeMetricJob(db database.DB) goroutine.BackgroundRoutine {
	job := &queueSizeJob{
		db: db,
		sizeGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "src",
			Name:      "telemetry_job_queue_size_total",
			Help:      "Current number of events waiting to be scraped.",
		}),
		throughputGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "src",
			Name:      "telemetry_job_max_throughput",
			Help:      "Currently configured maximum throughput per second.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("analytics.event-log-export-metrics"),
		goroutine.WithDescription("event logs export backlog metrics"),
		goroutine.WithInterval(time.Minute*5),
	)
}

type queueSizeJob struct {
	db              database.DB
	sizeGauge       prometheus.Gauge
	throughputGauge prometheus.Gauge
}

func (j *queueSizeJob) Handle(ctx context.Context) error {
	bookmarkStore := newBookmarkStore(j.db)
	bookmark, err := bookmarkStore.GetBookmark(ctx)
	if err != nil {
		return errors.Wrap(err, "queueSizeJob.GetBookmark")
	}

	store := basestore.NewWithHandle(j.db.Handle())
	val, err := basestore.ScanInt(store.QueryRow(ctx, sqlf.Sprintf("select count(*) from event_logs where id > %d and name in (select event_name from event_logs_export_allowlist);", bookmark)))
	if err != nil {
		return errors.Wrap(err, "queueSizeJob.GetCount")
	}
	j.sizeGauge.Set(float64(val))

	batchSize := getBatchSize()
	throughput := float64(batchSize) / float64(JobCooldownDuration/time.Second)
	j.throughputGauge.Set(throughput)

	return nil
}

func newBackgroundTelemetryJob(logger log.Logger, db database.DB) goroutine.BackgroundRoutine {
	observationCtx := observation.NewContext(log.NoOp())
	handlerMetrics := newHandlerMetrics(observationCtx)
	th := newTelemetryHandler(logger, db.EventLogs(), db.UserEmails(), db.GlobalState(), newBookmarkStore(db), sendEvents, handlerMetrics)
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		th,
		goroutine.WithName("analytics.telemetry-export"),
		goroutine.WithDescription("event logs telemetry sender"),
		goroutine.WithInterval(JobCooldownDuration),
		goroutine.WithOperation(handlerMetrics.handler),
	)
}

type sendEventsCallbackFunc func(ctx context.Context, event []*database.Event, config topicConfig, metadata instanceMetadata) error

func newHandlerMetrics(observationCtx *observation.Context) *handlerMetrics {
	redM := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"telemetry_job",
		metrics.WithLabels("op"),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("telemetry_job.telemetry_handler.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redM,
		})
	}
	return &handlerMetrics{
		sendEvents:  op("SendEvents"),
		fetchEvents: op("FetchEvents"),
		handler:     op("Handler"),
	}
}

type handlerMetrics struct {
	handler     *observation.Operation
	sendEvents  *observation.Operation
	fetchEvents *observation.Operation
}

type telemetryHandler struct {
	logger             log.Logger
	eventLogStore      database.EventLogStore
	globalStateStore   database.GlobalStateStore
	userEmailsStore    database.UserEmailsStore
	bookmarkStore      bookmarkStore
	sendEventsCallback sendEventsCallbackFunc
	metrics            *handlerMetrics
}

func newTelemetryHandler(logger log.Logger, store database.EventLogStore, userEmailsStore database.UserEmailsStore, globalStateStore database.GlobalStateStore, bookmarkStore bookmarkStore, sendEventsCallback sendEventsCallbackFunc, metrics *handlerMetrics) *telemetryHandler {
	return &telemetryHandler{
		logger:             logger,
		eventLogStore:      store,
		sendEventsCallback: sendEventsCallback,
		globalStateStore:   globalStateStore,
		userEmailsStore:    userEmailsStore,
		bookmarkStore:      bookmarkStore,
		metrics:            metrics,
	}
}

var disabledErr = errors.New("Usage telemetry export is disabled, but the background job is attempting to execute. This means the configuration was disabled without restarting the worker service. This job is aborting, and no telemetry will be exported.")

const (
	MaxEventsCountDefault = 1000
	JobCooldownDuration   = time.Second * 60
)

func (t *telemetryHandler) Handle(ctx context.Context) (err error) {
	if !isEnabled() {
		return disabledErr
	}
	topicConfig, err := getTopicConfig()
	if err != nil {
		return errors.Wrap(err, "getTopicConfig")
	}

	instanceMetadata, err := getInstanceMetadata(ctx, t.globalStateStore, t.userEmailsStore)
	if err != nil {
		return errors.Wrap(err, "getInstanceMetadata")
	}

	batchSize := getBatchSize()

	bookmark, err := t.bookmarkStore.GetBookmark(ctx)
	if err != nil {
		return errors.Wrap(err, "GetBookmark")
	}
	t.logger.Info("fetching events from bookmark", log.Int("bookmark_id", bookmark))

	all, err := fetchEvents(ctx, bookmark, batchSize, t.eventLogStore, t.metrics)
	if err != nil {
		return errors.Wrap(err, "fetchEvents")
	}
	if len(all) == 0 {
		return nil
	}

	maxId := int(all[len(all)-1].ID)
	t.logger.Info("telemetryHandler executed", log.Int("event count", len(all)), log.Int("maxId", maxId))

	err = sendBatch(ctx, all, topicConfig, instanceMetadata, t.metrics, t.sendEventsCallback)
	if err != nil {
		return errors.Wrap(err, "sendBatch")
	}

	return t.bookmarkStore.UpdateBookmark(ctx, maxId)
}

// sendBatch wraps the send events callback in a metric
func sendBatch(ctx context.Context, events []*database.Event, topicConfig topicConfig, metadata instanceMetadata, metrics *handlerMetrics, callback sendEventsCallbackFunc) (err error) {
	ctx, _, endObservation := metrics.sendEvents.With(ctx, &err, observation.Args{})
	sentCount := 0
	defer func() { endObservation(float64(sentCount), observation.Args{}) }()

	err = callback(ctx, events, topicConfig, metadata)
	if err != nil {
		return err
	}
	sentCount = len(events)
	return nil
}

// fetchEvents wraps the event data fetch in a metric
func fetchEvents(ctx context.Context, bookmark, batchSize int, eventLogStore database.EventLogStore, metrics *handlerMetrics) (results []*database.Event, err error) {
	ctx, _, endObservation := metrics.fetchEvents.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(results)), observation.Args{}) }()

	return eventLogStore.ListExportableEvents(ctx, bookmark, batchSize)
}

// This package level client is to prevent race conditions when mocking this configuration in tests.
var confClient = conf.DefaultClient()

func isEnabled() bool {
	return enabled
}

func getBatchSize() int {
	config := confClient.Get()
	if config == nil || config.ExportUsageTelemetry == nil || config.ExportUsageTelemetry.BatchSize <= 0 {
		return MaxEventsCountDefault
	}
	return config.ExportUsageTelemetry.BatchSize
}

type topicConfig struct {
	projectName string
	topicName   string
}

func getTopicConfig() (topicConfig, error) {
	var config topicConfig

	config.topicName = topicName
	if config.topicName == "" {
		return config, errors.New("missing topic name to export usage data")
	}
	config.projectName = projectName
	if config.projectName == "" {
		return config, errors.New("missing project name to export usage data")
	}
	return config, nil
}

const (
	enabledEnvVar     = "EXPORT_USAGE_DATA_ENABLED"
	topicNameEnvVar   = "EXPORT_USAGE_DATA_TOPIC_NAME"
	projectNameEnvVar = "EXPORT_USAGE_DATA_TOPIC_PROJECT"
)

var (
	enabled, _  = strconv.ParseBool(env.Get(enabledEnvVar, "false", "Export usage data from this Sourcegraph instance to centralized Sourcegraph analytics (requires restart)."))
	topicName   = env.Get(topicNameEnvVar, "", "GCP pubsub topic name for event level data usage exporter")
	projectName = env.Get(projectNameEnvVar, "", "GCP project name for pubsub topic for event level data usage exporter")
)

func emptyIfNil(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildBigQueryObject(event *database.Event, metadata *instanceMetadata) *bigQueryEvent {
	return &bigQueryEvent{
		EventName:         event.Name,
		UserID:            int(event.UserID),
		AnonymousUserID:   event.AnonymousUserID,
		URL:               "", // omitting URL intentionally
		Source:            event.Source,
		Timestamp:         event.Timestamp.Format(time.RFC3339),
		PublicArgument:    string(event.PublicArgument),
		Version:           event.Version, // sending event Version since these events could be scraped from the past
		SiteID:            metadata.SiteID,
		LicenseKey:        metadata.LicenseKey,
		DeployType:        metadata.DeployType,
		InitialAdminEmail: metadata.InitialAdminEmail,
		FeatureFlags:      string(event.EvaluatedFlagSet.Json()),
		CohortID:          event.CohortID,
		FirstSourceURL:    emptyIfNil(event.FirstSourceURL),
		LastSourceURL:     emptyIfNil(event.LastSourceURL),
		Referrer:          emptyIfNil(event.Referrer),
		DeviceID:          event.DeviceID,
		InsertID:          event.InsertID,
	}
}

func sendEvents(ctx context.Context, events []*database.Event, config topicConfig, metadata instanceMetadata) error {
	client, err := pubsub.NewClient(ctx, config.projectName)
	if err != nil {
		return errors.Wrap(err, "pubsub.NewClient")
	}
	defer client.Close()

	var toSend []*bigQueryEvent
	for _, event := range events {
		pubsubEvent := buildBigQueryObject(event, &metadata)
		toSend = append(toSend, pubsubEvent)
	}

	marshal, err := json.Marshal(toSend)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	topic := client.Topic(config.topicName)
	defer topic.Stop()
	masg := &pubsub.Message{
		Data: marshal,
	}
	result := topic.Publish(ctx, masg)
	_, err = result.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "result.Get")
	}

	return nil
}

type bigQueryEvent struct {
	SiteID            string  `json:"site_id"`
	LicenseKey        string  `json:"license_key"`
	InitialAdminEmail string  `json:"initial_admin_email"`
	DeployType        string  `json:"deploy_type"`
	EventName         string  `json:"name"`
	URL               string  `json:"url"`
	AnonymousUserID   string  `json:"anonymous_user_id"`
	FirstSourceURL    string  `json:"first_source_url"`
	LastSourceURL     string  `json:"last_source_url"`
	UserID            int     `json:"user_id"`
	Source            string  `json:"source"`
	Timestamp         string  `json:"timestamp"`
	Version           string  `json:"Version"`
	FeatureFlags      string  `json:"feature_flags"`
	CohortID          *string `json:"cohort_id,omitempty"`
	Referrer          string  `json:"referrer,omitempty"`
	PublicArgument    string  `json:"public_argument"`
	DeviceID          *string `json:"device_id,omitempty"`
	InsertID          *string `json:"insert_id,omitempty"`
}

type instanceMetadata struct {
	DeployType        string
	Version           string
	SiteID            string
	LicenseKey        string
	InitialAdminEmail string
}

func getInstanceMetadata(ctx context.Context, stateStore database.GlobalStateStore, userEmailsStore database.UserEmailsStore) (instanceMetadata, error) {
	siteId, err := getSiteId(ctx, stateStore)
	if err != nil {
		return instanceMetadata{}, errors.Wrap(err, "getInstanceMetadata.getSiteId")
	}

	initialAdminEmail, err := getInitialAdminEmail(ctx, userEmailsStore)
	if err != nil {
		return instanceMetadata{}, errors.Wrap(err, "getInstanceMetadata.getInitialAdminEmail")
	}

	return instanceMetadata{
		DeployType:        deploy.Type(),
		Version:           version.Version(),
		SiteID:            siteId,
		LicenseKey:        confClient.Get().LicenseKey,
		InitialAdminEmail: initialAdminEmail,
	}, nil
}

func getSiteId(ctx context.Context, store database.GlobalStateStore) (string, error) {
	state, err := store.Get(ctx)
	if err != nil {
		return "", err
	}
	return state.SiteID, nil
}

func getInitialAdminEmail(ctx context.Context, store database.UserEmailsStore) (string, error) {
	info, _, err := store.GetInitialSiteAdminInfo(ctx)
	if err != nil {
		return "", err
	}
	return info, nil
}

type bmStore struct {
	*basestore.Store
}

func newBookmarkStore(db database.DB) bookmarkStore {
	return &bmStore{Store: basestore.NewWithHandle(db.Handle())}
}

type bookmarkStore interface {
	GetBookmark(ctx context.Context) (int, error)
	UpdateBookmark(ctx context.Context, val int) error
}

func (s *bmStore) GetBookmark(ctx context.Context) (_ int, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	val, found, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf("select bookmark_id from event_logs_scrape_state order by id limit 1;")))
	if err != nil {
		return 0, err
	}
	if !found {
		// generate a row and return the value
		return basestore.ScanInt(tx.QueryRow(ctx, sqlf.Sprintf("INSERT INTO event_logs_scrape_state (bookmark_id) SELECT MAX(id) FROM event_logs RETURNING bookmark_id;")))
	}
	return val, err
}

func (s *bmStore) UpdateBookmark(ctx context.Context, val int) error {
	return s.Exec(ctx, sqlf.Sprintf("UPDATE event_logs_scrape_state SET bookmark_id = %S WHERE id = (SELECT id FROM event_logs_scrape_state ORDER BY id LIMIT 1);", val))
}
