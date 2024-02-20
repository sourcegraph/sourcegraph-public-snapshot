package events

import (
	"context"
	"encoding/json"

	googlepubsub "cloud.google.com/go/pubsub"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cockroachdb/redact"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Publisher struct {
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

func NewPublisherForStream(eventsTopic pubsub.TopicPublisher, metadata *telemetrygatewayv1.RecordEventsRequestMetadata, opts PublishStreamOptions) (*Publisher, error) {
	metadataJSON, err := protojson.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling metadata")
	}
	if opts.ConcurrencyLimit <= 0 {
		opts.ConcurrencyLimit = 250
	}
	return &Publisher{
		topic:        eventsTopic,
		opts:         opts,
		metadataJSON: metadataJSON,
	}, nil
}

type PublishEventResult struct {
	// EventID is the ID of the event that was published.
	EventID string
	// PublishError, if non-nil, indicates an error occurred publishing the event.
	PublishError error
}

// Publish emits all events concurrently, up to 100 at a time for each call.
func (p *Publisher) Publish(ctx context.Context, events []*telemetrygatewayv1.Event) []PublishEventResult {
	wg := pool.NewWithResults[PublishEventResult]().
		WithMaxGoroutines(p.opts.ConcurrencyLimit) // limit each batch to some degree

	for _, event := range events {
		event := event // capture range variable :(

		doPublish := func(event *telemetrygatewayv1.Event) error {
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
			// TODO: We can't error forever, as the instance will keep trying to
			// deliver this event - eventually we may just need to take an action,
			// then pretend the event succeeded. Pending decision from data team
			// on what to do with these: https://sourcegraph.slack.com/archives/CN4FC7XT4/p1707986514302069
			if len(payload) >= googlepubsub.MaxPublishRequestBytes {
				return errors.Newf("event %s/%s is oversized (ID: %s, size: %s)",
					// Mark values as safe for cockroachdb Sentry reporting
					redact.Safe(event.GetFeature()),
					redact.Safe(event.GetAction()),
					redact.Safe(event.GetId()),
					len(payload))
			}

			// Publish a single message in each callback to manage concurrency
			// ourselves, and attach attributes for ease of routing the pub/sub
			// message.
			if err := p.topic.PublishMessage(ctx, payload, extractPubSubAttributes(event)); err != nil {
				// Try to record the cancel cause as the primary error in case
				// one is recorded.
				if cancelCause := context.Cause(ctx); cancelCause != nil {
					return errors.Wrapf(cancelCause, "%s: interrupted event publish",
						err.Error())
				}
				return errors.Wrap(err, "publishing event")
			}

			return nil
		}

		wg.Go(func() PublishEventResult {
			return PublishEventResult{
				EventID:      event.GetId(),
				PublishError: doPublish(event),
			}
		})
	}
	return wg.Wait()
}
