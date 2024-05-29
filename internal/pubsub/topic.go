// Package pubsub is a lightweight wrapper of the GCP Pub/Sub functionality.
package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TopicClient is a Pub/Sub client that bound to a topic.
type TopicClient interface {
	// TopicPublisher is the interface most callsites will use, which only
	// includes publishing entrypoints. When propagating TopicClient as a
	// dependency, prefer to use the smaller TopicPublisher interface.
	TopicPublisher

	// Ping checks if the connection to the topic is valid.
	Ping(ctx context.Context) error
	// Stop stops the topic publishing channel. The client should not be used after
	// calling Stop.
	Stop(ctx context.Context) error
}

// TopicPublisher is a Pub/Sub publisher bound to a topic.
type TopicPublisher interface {
	// Publish publishes messages and waits for all the results synchronously.
	// It returns the first error encountered or nil if all succeeded. To collect
	// individual errors, call Publish with only 1 message, or use PublishMessage.
	Publish(ctx context.Context, messages ...[]byte) error
	// PublishMessage publishes a single message with attributes and waits for
	// the result synchronously.
	PublishMessage(ctx context.Context, message []byte, attributes map[string]string) error
}

var (
	defaultProjectID       = env.Get("PUBSUB_PROJECT_ID", "", "The project ID of the Pub/Sub.")
	defaultCredentialsFile = env.Get("PUBSUB_CREDENTIALS_FILE", "", "The credentials file of the Pub/Sub project.")
)

// TopicClient is a Pub/Sub client that bound to a topic.
type topicClient struct {
	topic *pubsub.Topic
}

// NewTopicClient creates a Pub/Sub client that bound to a topic of the given
// project.
func NewTopicClient(projectID, topicID string, opts ...option.ClientOption) (TopicClient, error) {
	client, err := pubsub.NewClient(context.Background(), projectID, opts...)
	if err != nil {
		return nil, errors.Errorf("create Pub/Sub client: %v", err)
	}
	return &topicClient{
		topic: client.Topic(topicID),
	}, nil
}

// NewDefaultTopicClient creates a Pub/Sub client that bound to a topic with
// default project ID and credentials file, whose values are read from the
// environment variables `PUBSUB_PROJECT_ID` and `PUBSUB_CREDENTIALS_FILE`
// respectively. It is OK to have empty value for credentials file if the client
// can be authenticated via other means against the target project.
func NewDefaultTopicClient(topicID string) (TopicClient, error) {
	return NewTopicClient(defaultProjectID, topicID, option.WithCredentialsFile(defaultCredentialsFile))
}

func (c *topicClient) Ping(ctx context.Context) error {
	exists, err := c.topic.Exists(ctx)
	if err != nil {
		return err
	} else if !exists {
		return errors.New("topic does not exist")
	}
	return nil
}

func (c *topicClient) Publish(ctx context.Context, messages ...[]byte) error {
	results := make([]*pubsub.PublishResult, 0, len(messages))
	for _, msg := range messages {
		results = append(results, c.topic.Publish(ctx, &pubsub.Message{Data: msg}))
	}
	for _, result := range results {
		if _, err := result.Get(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *topicClient) PublishMessage(ctx context.Context, message []byte, attributes map[string]string) error {
	r := c.topic.Publish(ctx, &pubsub.Message{Data: message, Attributes: attributes})
	if _, err := r.Get(ctx); err != nil {
		return err
	}
	return nil
}

func (c *topicClient) Stop(context.Context) error {
	c.topic.Stop()
	return nil
}

// NewNoopTopicClient creates a no-op Pub/Sub client that does nothing on any
// method call. This is useful as a stub implementation of the TopicClient.
func NewNoopTopicClient() TopicClient {
	return &noopTopicClient{}
}

type noopTopicClient struct{}

func (c *noopTopicClient) Ping(context.Context) error               { return nil }
func (c *noopTopicClient) Publish(context.Context, ...[]byte) error { return nil }
func (c *noopTopicClient) PublishMessage(context.Context, []byte, map[string]string) error {
	return nil
}
func (c *noopTopicClient) Stop(context.Context) error { return nil }

// NewLoggingTopicClient creates a Pub/Sub client that just logs all messages,
// and does nothing otherwise. This is also a useful stub implementation of the
// TopicClient for testing/debugging purposes.
//
// Log entries are generated at debug level.
func NewLoggingTopicClient(logger log.Logger) TopicClient {
	return &loggingTopicClient{
		logger: logger.Scoped("pubsub"),
	}
}

type loggingTopicClient struct {
	logger log.Logger
	noopTopicClient
}

func (c *loggingTopicClient) Publish(ctx context.Context, messages ...[]byte) error {
	l := trace.Logger(ctx, c.logger)
	for _, m := range messages {
		l.Debug("Publish", log.String("message", string(m)))
	}
	return nil
}

func (c *loggingTopicClient) PublishMessage(ctx context.Context, message []byte, attributes map[string]string) error {
	attributesFields := make([]log.Field, 0, len(attributes))
	for k, v := range attributes {
		attributesFields = append(attributesFields, log.String(k, v))
	}
	trace.Logger(ctx, c.logger).Debug("Publish",
		log.String("message", string(message)),
		log.Object("attributes", attributesFields...))
	return nil
}
