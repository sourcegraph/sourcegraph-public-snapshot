package events

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/telemetry-gateway/shared/events"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TopicConfig struct {
	ProjectName string
	TopicName   string
}

type EventProxy struct {
	client *pubsub.Client
	config TopicConfig
}

func NewEventProxy(config TopicConfig) (*EventProxy, error) {
	client, err := pubsub.NewClient(context.Background(), config.ProjectName)
	if err != nil {
		return nil, errors.Wrap(err, "pubsub.NewClient")
	}
	return &EventProxy{client: client, config: config}, nil
}

func (e *EventProxy) SendEvents(ctx context.Context, request events.TelemetryGatewayProxyRequest) error {
	marshal, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	topic := e.client.Topic(e.config.TopicName)
	defer topic.Stop()
	msg := &pubsub.Message{
		Data: marshal,
	}
	result := topic.Publish(ctx, msg)
	_, err = result.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "result.Get")
	}

	return nil
}

type EventSender interface {
	SendEvents(ctx context.Context, request events.TelemetryGatewayProxyRequest) error
}

type FakeEventSender struct {
}

func (f *FakeEventSender) SendEvents(ctx context.Context, request events.TelemetryGatewayProxyRequest) error {
	jsonify, _ := json.Marshal(request)
	fmt.Printf("%v\n", string(jsonify))
	return nil
}
