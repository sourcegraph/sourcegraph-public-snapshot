package events

import (
	"context"
	"encoding/json"
	"strings"

	googlepubsub "cloud.google.com/go/pubsub"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cockroachdb/redact"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Publisher struct {
	logger log.Logger
	source string

	topic pubsub.TopicPublisher
	opts  PublishStreamOptions

	metadataJSON json.RawMessage
}

type PublishStreamOptions struct {
	// ConcurrencyLimit sets the maximum number of concurrent publishes for
	// a stream.
	ConcurrencyLimit int
	// MessageSizeHistogram, if provided, records the size of message payloads
	// published to the events topic.
	MessageSizeHistogram metric.Int64Histogram
}

func NewPublisherForStream(
	logger log.Logger,
	eventsTopic pubsub.TopicPublisher,
	metadata *telemetrygatewayv1.RecordEventsRequestMetadata,
	opts PublishStreamOptions,
) (*Publisher, error) {
	metadataJSON, err := protojson.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling metadata")
	}
	if opts.ConcurrencyLimit <= 0 {
		opts.ConcurrencyLimit = 250
	}

	var source string
	switch identifier := metadata.GetIdentifier(); identifier.GetIdentifier().(type) {
	case *telemetrygatewayv1.Identifier_LicensedInstance:
		source = "licensed_instance"
	case *telemetrygatewayv1.Identifier_UnlicensedInstance:
		source = "unlicensed_instance"
	case *telemetrygatewayv1.Identifier_ManagedService:
		// Is a trusted client, so use the service ID directly as the source
		source = identifier.GetManagedService().ServiceId
	default:
		source = "unknown"
	}

	return &Publisher{
		logger:       logger.With(log.String("source", source)),
		source:       source,
		topic:        eventsTopic,
		opts:         opts,
		metadataJSON: metadataJSON,
	}, nil
}

// GetSourceName returns a name inferred from metadata provided to
// NewPublisherForStream, for use as a metric label. It is safe to call on a nil
// publisher.
func (p *Publisher) GetSourceName() string {
	if p == nil {
		return "invalid"
	}
	return p.source
}

type PublishEventResult struct {
	// EventID is the ID of the event that was published.
	EventID string
	// EventFeature is the feature of the event that was published.
	EventFeature string
	// EventAction is the action of the event that was published.
	EventAction string
	// EventSource is a string representation of source of the event, as reported
	// at recording time, in the default proto string format, e.g:
	//
	//   server:{version:"..."}  client:{name:"..."}
	EventSource string
	// PublishError, if non-nil, indicates an error occurred publishing the event.
	PublishError error
}

// NewPublishEventResult returns a PublishEventResult for the given event and error.
// Should only be used internally or in testing.
func NewPublishEventResult(event *telemetrygatewayv1.Event, err error) PublishEventResult {
	return PublishEventResult{
		EventID:      event.GetId(),
		EventFeature: event.GetFeature(),
		EventAction:  event.GetAction(),
		EventSource:  event.GetSource().String(),
		PublishError: err,
	}
}

// Publish emits all events concurrently, up to 100 at a time for each call.
func (p *Publisher) Publish(ctx context.Context, events []*telemetrygatewayv1.Event) []PublishEventResult {
	wg := pool.NewWithResults[PublishEventResult]().
		WithMaxGoroutines(p.opts.ConcurrencyLimit) // limit each batch to some degree

	for _, event := range events {
		event := event // capture range variable :(

		doPublish := func(event *telemetrygatewayv1.Event) error {
			// Ensure the most important fields are in place
			if event.Id == "" {
				return errors.New("event ID is required")
			}
			if event.Feature == "" {
				return errors.New("event feature is required")
			}
			if event.Action == "" {
				return errors.New("event action is required")
			}
			if event.Timestamp == nil {
				return errors.New("event timestamp is required")
			}

			// Render JSON format for publishing
			eventJSON, err := protojson.Marshal(event)
			if err != nil {
				return errors.Wrap(err, "marshalling event")
			}

			// Join our raw JSON payloads into a single message
			payload, err := json.Marshal(map[string]json.RawMessage{
				"metadata": p.metadataJSON,
				"event":    json.RawMessage(eventJSON),
			})
			if err != nil {
				return errors.Wrap(err, "marshalling event payload")
			}

			if p.opts.MessageSizeHistogram != nil {
				p.opts.MessageSizeHistogram.Record(ctx, int64(len(payload)))
			}

			// If the payload is obviously oversized, don't bother publishing it
			// - record it with some additional details instead
			//
			// We can't error forever, as the instance will keep trying to deliver
			// this event - for now, we just pretend the event succeeded, and log
			// some diagnostics.
			//
			// TODO: Maybe we can merge this with how we handle errors in
			// summarizePublishEventsResults - for now, we stick with this
			// special handling for extra visibility.
			if len(payload) >= googlepubsub.MaxPublishRequestBytes {
				trace.Logger(ctx, p.logger).Error("discarding oversized event",
					log.Error(errors.Newf("event %s/%s is oversized",
						// Mark values as safe for cockroachdb Sentry reporting
						// Include this metadata in the actual error message so
						// that each occurrence is treated as its own Sentry
						// alert
						redact.Safe(event.GetFeature()),
						redact.Safe(event.GetAction()))),
					log.String("eventID", event.GetId()),
					log.String("eventSource", event.GetSource().String()),
					log.Int("size", len(payload)),
					// Record a section of the event content for diagnostics.
					// Keep in mind the size of Context objects in Sentry:
					// https://develop.sentry.dev/sdk/data-handling/#variable-size
					// And GCP logging limits: https://cloud.google.com/logging/quotas
					log.String("eventSnippet", strings.ToValidUTF8(string(eventJSON[:512]), "�")))
				// We must return nil, pretending the publish succeeded, so that
				// the client stops attempting to publish an event that will
				// never succeed.
				return nil
			}

			// Publish a single message in each callback to manage concurrency
			// ourselves, and attach attributes for ease of routing the pub/sub
			// message.
			if err := p.topic.PublishMessage(ctx, payload, extractPubSubAttributes(p.source, event)); err != nil {
				// Explicitly record the cancel cause if one is provided.
				if cancelCause := context.Cause(ctx); cancelCause != nil {
					return errors.Wrapf(err, "interrupted event publish, cause: %s",
						errors.Safe(cancelCause.Error()))
				}
				return errors.Wrap(err, "publishing event")
			}

			return nil
		}

		wg.Go(func() PublishEventResult {
			return NewPublishEventResult(event, doPublish(event))
		})
	}
	return wg.Wait()
}
