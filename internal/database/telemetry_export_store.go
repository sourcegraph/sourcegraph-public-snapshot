package database

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/sensitivemetadataallowlist"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

var counterQueuedEvents = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "telemetry_export_store",
	Name:      "queued_events",
	Help:      "Events added to the telemetry export queue",
}, []string{"failed"})

type TelemetryEventsExportQueueStore interface {
	basestore.ShareableStore

	// QueueForExport caches a set of events for later export. It is currently
	// feature-flagged, such that if the flag is not enabled for the given
	// context, we do not cache the event for export.
	//
	// It does NOT respect context cancellation, as it is assumed that we never
	// drop events once we attempt to queue it for export.
	//
	// 🚨 SECURITY: The implementation strips out sensitive contents from events
	// that are not in sensitivemetadataallowlist.AllowedEventTypes().
	//
	// The implementation may also drop events based on the export policy
	// allowed on the instance - see licensing.TelemetryEventsExportMode
	QueueForExport(context.Context, []*telemetrygatewayv1.Event) error

	// ListForExport returns the cached events that should be exported next. All
	// events returned should be exported.
	ListForExport(ctx context.Context, limit int) ([]*telemetrygatewayv1.Event, error)

	// MarkAsExported marks all events in the set of IDs as exported.
	MarkAsExported(ctx context.Context, eventIDs []string) error

	// DeletedExported deletes all events exported before the given timestamp,
	// returning the number of affected events.
	DeletedExported(ctx context.Context, before time.Time) (int64, error)

	TelemetryEventsExportQueueDiagnosticsStore
}

// TelemetryEventsExportQueueDiagnosticsStore is a read-only subset of
// TelemetryEventsExportQueueStore for diagnostics endpoints and helpers.
type TelemetryEventsExportQueueDiagnosticsStore interface {
	// CountUnexported returns the number of events not yet exported. The time
	// returned indicates the timestamp of the oldest event - it may be a zero
	// value if there are no events awaiting export.
	CountUnexported(ctx context.Context) (int64, time.Time, error)

	// CountRecentlyExported returns the number of events recently exported.
	// Data retention depends on TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION.
	CountRecentlyExported(ctx context.Context) (int64, error)

	// ListRecentlyExported returns the most recently exported events before
	// the cursor timestamp.
	// Data retention depends on TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION.
	ListRecentlyExported(ctx context.Context, first int, before *time.Time) ([]ExportedTelemetryEvent, error)
}

func TelemetryEventsExportQueueWith(logger log.Logger, other basestore.ShareableStore) TelemetryEventsExportQueueStore {
	return &telemetryEventsExportQueueStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

type telemetryEventsExportQueueStore struct {
	logger log.Logger
	*basestore.Store

	// mockExportMode can be set in tests to imitate a particular export mode.
	// It can be configured by casting the store into the
	// MockExportModeSetterTelemetryEventsExportQueueStore interface.
	mockExportMode *licensing.TelemetryEventsExportMode
}

// getExportMode calls licensing.GetTelemetryEventsExportMode or returns a mock
// export mode if configured in testing.
func (s *telemetryEventsExportQueueStore) getExportMode() licensing.TelemetryEventsExportMode {
	if s.mockExportMode != nil {
		return *s.mockExportMode
	}
	return licensing.GetTelemetryEventsExportMode(conf.DefaultClient())
}

// See interface docstring.
func (s *telemetryEventsExportQueueStore) QueueForExport(ctx context.Context, events []*telemetrygatewayv1.Event) (err error) {
	var tr trace.Trace
	tr, ctx = trace.New(ctx, "telemetryevents.QueueForExport",
		// actually queued events may be different - final attribute is added later
		attribute.Int("submitted-events", len(events)))
	defer tr.EndWithErr(&err)

	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Respect export mode carefully.
	switch s.getExportMode() {
	case licensing.TelemetryEventsExportAll:
		tr.SetAttributes(attribute.String("export-mode", "enabled"))

	case licensing.TelemetryEventsExportCodyOnly:
		// 🚨 SECURITY: Only export Cody-related events in this mode - drop
		// everything else from the set.
		var dropped int
		filtered := make([]*telemetrygatewayv1.Event, 0, len(events))
		for _, e := range events {
			if e.Feature == "cody" || strings.HasPrefix(e.Feature, "cody.") {
				filtered = append(filtered, e)
			} else {
				dropped += 1
			}
		}
		events = filtered // queued filtered for export
		tr.SetAttributes(
			attribute.String("export-mode", "cody-only"),
			attribute.Int("dropped-events", dropped))

	case licensing.TelemetryEventsExportDisabled:
		// 🚨 SECURITY: Do not export any events in this mode.
		tr.SetAttributes(attribute.String("export-mode", "disabled"))
		return nil // no-op

	default:
		return errors.Newf("telemetry: unknown export mode %+v", s.getExportMode())
	}

	if len(events) == 0 {
		return nil
	}

	// Create a cancel-free context to avoid interrupting the insert when
	// the parent context is cancelled, and add our own timeout on the insert
	// to make sure things don't get stuck in an unbounded manner.
	insertCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	defer cancel()

	err = batch.InsertValues(
		insertCtx,
		s.Handle(),
		"telemetry_events_export_queue",
		batch.MaxNumPostgresParameters,
		[]string{
			"id",
			"timestamp",
			"payload_pb",
		},
		insertTelemetryEventsChannel(logger, events))

	// Record results
	counterQueuedEvents.
		WithLabelValues(strconv.FormatBool(err != nil)).
		Add(float64(len(events)))
	if err == nil {
		tr.SetAttributes(attribute.Int("queued-events", len(events)))
	}

	return err
}

func insertTelemetryEventsChannel(logger log.Logger, events []*telemetrygatewayv1.Event) <-chan []any {
	ch := make(chan []any, len(events))

	go func() {
		defer close(ch)

		sensitiveAllowlist := sensitivemetadataallowlist.AllowedEventTypes()
		for _, event := range events {
			// 🚨 SECURITY: Apply sensitive data redaction of the payload.
			// Redaction mutates the payload so we should make a copy.
			event := proto.Clone(event).(*telemetrygatewayv1.Event)
			sensitiveAllowlist.Redact(event)

			payloadPB, err := proto.Marshal(event)
			if err != nil {
				logger.Error("failed to marshal telemetry event",
					log.String("event.feature", event.GetFeature()),
					log.String("event.action", event.GetAction()),
					log.String("event.source.client.name", event.GetSource().GetClient().GetName()),
					log.String("event.source.client.version", event.GetSource().GetClient().GetVersion()),
					log.Error(err))
				continue
			}
			ch <- []any{
				event.Id,                 // id
				event.Timestamp.AsTime(), // timestamp
				payloadPB,                // payload_pb
			}
		}
	}()

	return ch
}

// See interface docstring.
func (s *telemetryEventsExportQueueStore) ListForExport(ctx context.Context, limit int) ([]*telemetrygatewayv1.Event, error) {
	var tr trace.Trace
	tr, ctx = trace.New(ctx, "telemetryevents.ListForExport",
		attribute.Int("limit", limit))
	defer tr.End()

	logger := trace.Logger(ctx, s.logger)

	rows, err := s.Store.Handle().QueryContext(ctx, `
		SELECT id, payload_pb
		FROM telemetry_events_export_queue
		WHERE exported_at IS NULL
		ORDER BY timestamp ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*telemetrygatewayv1.Event, 0, limit)
	for rows.Next() {
		var id string
		var payloadPB []byte
		err := rows.Scan(&id, &payloadPB)
		if err != nil {
			return nil, err
		}

		event := &telemetrygatewayv1.Event{}
		if err := proto.Unmarshal(payloadPB, event); err != nil {
			tr.RecordError(err)
			logger.Error("failed to unmarshal telemetry event payload",
				log.String("id", id),
				log.Error(err))
			// Not fatal, just ignore this event for now, leaving it in DB for
			// investigation.
			continue
		}

		events = append(events, event)
	}
	tr.SetAttributes(attribute.Int("events", len(events)))
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// See interface docstring.
func (s *telemetryEventsExportQueueStore) MarkAsExported(ctx context.Context, eventIDs []string) error {
	if _, err := s.Store.Handle().ExecContext(ctx, `
		UPDATE telemetry_events_export_queue
		SET exported_at = NOW()
		WHERE id = ANY($1);
	`, pq.Array(eventIDs)); err != nil {
		return errors.Wrap(err, "failed to mark events as exported")
	}
	return nil
}

func (s *telemetryEventsExportQueueStore) DeletedExported(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.Store.Handle().ExecContext(ctx, `
	DELETE FROM telemetry_events_export_queue
	WHERE
		exported_at IS NOT NULL
		AND exported_at < $1;
`, before)
	if err != nil {
		return 0, errors.Wrap(err, "failed to mark events as exported")
	}
	return result.RowsAffected()
}

func (s *telemetryEventsExportQueueStore) CountUnexported(ctx context.Context) (int64, time.Time, error) {
	var count int64
	var oldest sql.NullTime
	err := s.Store.Handle().QueryRowContext(ctx, `
	SELECT COUNT(*), MIN(timestamp)
	FROM telemetry_events_export_queue
	WHERE exported_at IS NULL;
	`).Scan(&count, &oldest)
	return count, oldest.Time, err
}

func (s *telemetryEventsExportQueueStore) CountRecentlyExported(ctx context.Context) (int64, error) {
	var count int64
	return count, s.Store.Handle().QueryRowContext(ctx, `
	SELECT COUNT(*)
	FROM telemetry_events_export_queue
	WHERE exported_at IS NOT NULL
	`).Scan(&count)
}

type ExportedTelemetryEvent struct {
	// ID is the unique ID of the event.
	ID string
	// ExportedAt is the timestamp when the event was marked as exported.
	ExportedAt time.Time
	// Timestamp of the payload, also used as the cursor for pagination.
	Timestamp time.Time
	// Payload is the data that was exported from the instance.
	Payload *telemetrygatewayv1.Event
}

func (s *telemetryEventsExportQueueStore) ListRecentlyExported(ctx context.Context, first int, before *time.Time) ([]ExportedTelemetryEvent, error) {
	var tr trace.Trace
	tr, ctx = trace.New(ctx, "telemetryevents.ListRecentlyExported")
	defer tr.End()

	logger := trace.Logger(ctx, s.logger)

	conds := []*sqlf.Query{
		sqlf.Sprintf("exported_at IS NOT NULL"),
	}
	if before != nil {
		// ListRecentlyExported is intended for reference, so we don't worry
		// about timestamp collisions with page size - we just strictly
		// get events older than the previous page.
		conds = append(conds, sqlf.Sprintf("timestamp < %s", before))
	}

	rows, err := s.Store.Query(ctx, sqlf.Sprintf(`
	SELECT
		id, exported_at, timestamp, payload_pb
	FROM telemetry_events_export_queue
	WHERE
		%s
	ORDER BY
		timestamp DESC -- get most recent entries first
	LIMIT %s`, sqlf.Join(conds, "AND"), first))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]ExportedTelemetryEvent, 0, first)
	for rows.Next() {
		var id string
		var exportedAt time.Time
		var timestamp time.Time
		var payloadPB []byte
		err := rows.Scan(&id, &exportedAt, &timestamp, &payloadPB)
		if err != nil {
			return nil, err
		}

		event := &telemetrygatewayv1.Event{}
		if err := proto.Unmarshal(payloadPB, event); err != nil {
			tr.RecordError(err)
			logger.Error("failed to unmarshal telemetry event payload",
				log.String("id", id),
				log.Error(err))
			// Not fatal, just ignore this event for now, leaving it in DB for
			// investigation.
			continue
		}

		events = append(events, ExportedTelemetryEvent{
			ID:         id,
			ExportedAt: exportedAt,
			Timestamp:  timestamp,
			Payload:    event,
		})
	}
	tr.SetAttributes(attribute.Int("events", len(events)))
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// MockExportModeSetterTelemetryEventsExportQueueStore can be cast from
// TelemetryEventsExportQueueStore for use in testing only.
//
// ⚠️ Use in tests only!
type MockExportModeSetterTelemetryEventsExportQueueStore interface {
	TelemetryEventsExportQueueStore

	// SetMockExportMode configures the store's mock export mode for use in testing.
	//
	// ⚠️ Use in tests only!
	SetMockExportMode(mode licensing.TelemetryEventsExportMode)
}

var _ MockExportModeSetterTelemetryEventsExportQueueStore = (*telemetryEventsExportQueueStore)(nil)

// See MockExportModeSetterTelemetryEventsExportQueueStore docstrings.
func (s *telemetryEventsExportQueueStore) SetMockExportMode(mode licensing.TelemetryEventsExportMode) {
	s.mockExportMode = &mode
}
