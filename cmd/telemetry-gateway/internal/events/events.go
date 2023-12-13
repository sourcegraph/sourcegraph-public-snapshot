package events

import (
	"context"
	"encoding/json"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Publisher struct {
	topic pubsub.TopicClient

	metadataJSON json.RawMessage
}

func NewPublisherForStream(eventsTopic pubsub.TopicClient, metadata *telemetrygatewayv1.RecordEventsRequestMetadata) (*Publisher, error) {
	metadataJSON, err := protojson.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling metadata")
	}
	return &Publisher{
		topic:        eventsTopic,
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
		WithMaxGoroutines(100) // limit each batch to some degree

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

			// Publish a single message in each callback to manage concurrency
			// ourselves, and
			if err := p.topic.PublishMessage(ctx, payload, map[string]string{
				"event.feature": event.Feature,
				"event.action":  event.Action,
				"event.hasPrivateMetadata": strconv.FormatBool(
					event.GetParameters().GetPrivateMetadata() != nil),
			}); err != nil {
				// Try to record the cancel cause in case one is recorded.
				if cancelCause := context.Cause(ctx); cancelCause != nil {
					return errors.Wrap(err, "interrupted event publish")
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
