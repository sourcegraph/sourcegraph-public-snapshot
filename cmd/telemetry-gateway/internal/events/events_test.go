package events_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/pubsub/pubsubtest"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestPublish(t *testing.T) {
	memTopic := pubsubtest.NewMemoryTopicClient()

	publisher, err := events.NewPublisherForStream(memTopic, &telemetrygatewayv1.RecordEventsRequestMetadata{})
	require.NoError(t, err)

	events := []*telemetrygatewayv1.Event{
		{Id: "1"},
		{Id: "2"},
		{Id: "3"},
		{Id: "4"},
		{Id: "5"},
	}

	// Collect sets of things we expect
	eventResults := make(map[string]bool)
	for _, e := range events {
		eventResults[e.Id] = false
	}
	publishedEvents := make(map[string]bool)
	for _, e := range events {
		publishedEvents[e.Id] = false
	}

	// Publish the events
	results := publisher.Publish(context.Background(), events)

	// Collect all the results we got
	for _, r := range results {
		eventResults[r.EventID] = true
	}

	// Collect all the messages we published
	for _, m := range memTopic.Messages {
		var payload map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(m, &payload))

		var event telemetrygatewayv1.Event
		require.NoError(t, json.Unmarshal(payload["event"], &event))
		publishedEvents[event.GetId()] = true
	}

	// Make our assertions - all events should be have results or be published
	t.Logf("results: %+v", eventResults)
	for eventID, found := range eventResults {
		if !found {
			t.Errorf("expected event result %q", eventID)
		}
	}
	t.Logf("published: %+v", publishedEvents)
	for eventID, published := range publishedEvents {
		if !published {
			t.Errorf("expected published event %q", eventID)
		}
	}
}
