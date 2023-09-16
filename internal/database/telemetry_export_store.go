package database

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FeatureFlagTelemetryExport enables telemetry export by allowing events to be
// queued for export via (TelemetryEventsExportQueueStore).QueueForExport
const FeatureFlagTelemetryExport = "telemetry-export"

type TelemetryEventsExportQueueStore interface {
	basestore.ShareableStore

	// QueueForExport caches a set of events for later export. It is currently
	// feature-flagged, such that if the flag is not enabled for the given
	// context, we do not cache the event for export.
	QueueForExport(context.Context, []*telemetrygatewayv1.Event) error

	// ListForExport returns the cached events that should be exported next. All
	// events returned should be exported.
	//
	// 🚨 SECURITY: Potentially sensitive parts of the payload are retained at
	// this stage. The caller is responsible for ensuring sensitive data is
	// stripped.
	ListForExport(ctx context.Context, limit int) ([]*telemetrygatewayv1.Event, error)

	// MarkAsExported marks all events in the set of IDs as exported.
	MarkAsExported(ctx context.Context, eventIDs []string) error

	// DeletedExported deletes all events exported before the given timestamp,
	// returning the number of affected events.
	DeletedExported(ctx context.Context, before time.Time) (int64, error)
}

func TelemetryEventsExportQueueWith(logger log.Logger, other basestore.ShareableStore) TelemetryEventsExportQueueStore {
	return &telemetryEventsExportQueueStore{
		logger:         logger,
		ShareableStore: other,
	}
}

type telemetryEventsExportQueueStore struct {
	logger log.Logger
	basestore.ShareableStore
}

// See interface docstring.
func (s *telemetryEventsExportQueueStore) QueueForExport(ctx context.Context, events []*telemetrygatewayv1.Event) error {
	var tr trace.Trace
	tr, ctx = trace.New(ctx, "telemetryevents.QueueForExport",
		attribute.Int("events", len(events)))
	defer tr.End()

	logger := trace.Logger(ctx, s.logger)

	if flags := featureflag.FromContext(ctx); flags == nil || !flags.GetBoolOr(FeatureFlagTelemetryExport, false) {
		tr.SetAttributes(attribute.Bool("enabled", false))
		return nil // no-op
	} else {
		tr.SetAttributes(attribute.Bool("enabled", true))
	}

	if len(events) == 0 {
		return nil
	}
	return batch.InsertValues(ctx,
		s.Handle(),
		"telemetry_events_export_queue",
		batch.MaxNumPostgresParameters,
		[]string{
			"id",
			"timestamp",
			"payload_pb",
		},
		insertChannel(logger, events))
}

func insertChannel(logger log.Logger, events []*telemetrygatewayv1.Event) <-chan []any {
	ch := make(chan []any, len(events))

	go func() {
		defer close(ch)

		for _, ev := range events {
			payloadPB, err := proto.Marshal(ev)
			if err != nil {
				logger.Error("failed to marshal telemetry event",
					log.String("event.feature", ev.GetFeature()),
					log.String("event.action", ev.GetAction()),
					log.String("event.source.client.name", ev.GetSource().GetClient().GetName()),
					log.String("event.source.client.version", ev.GetSource().GetClient().GetVersion()),
					log.Error(err))
				continue
			}
			ch <- []any{
				ev.Id,                 // id
				ev.Timestamp.AsTime(), // timestamp
				payloadPB,             // payload_pb
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

	rows, err := s.ShareableStore.Handle().QueryContext(ctx, `
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
		var event telemetrygatewayv1.Event
		if err := proto.Unmarshal(payloadPB, &event); err != nil {
			tr.RecordError(err)
			logger.Error("failed to unmarshal telemetry event payload",
				log.String("id", id),
				log.Error(err))
			// Not fatal, just ignore this event for now, leaving it in DB for
			// investigation.
			continue
		}
		events = append(events, &event)
	}
	tr.SetAttributes(attribute.Int("events", len(events)))
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// See interface docstring.
func (s *telemetryEventsExportQueueStore) MarkAsExported(ctx context.Context, eventIDs []string) error {
	if _, err := s.ShareableStore.Handle().ExecContext(ctx, `
		UPDATE telemetry_events_export_queue
		SET exported_at = NOW()
		WHERE id = ANY($1);
	`, pq.Array(eventIDs)); err != nil {
		return errors.Wrap(err, "failed to mark events as exported")
	}
	return nil
}

func (s *telemetryEventsExportQueueStore) DeletedExported(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.ShareableStore.Handle().ExecContext(ctx, `
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
