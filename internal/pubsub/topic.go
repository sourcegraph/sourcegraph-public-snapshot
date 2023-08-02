// Package pubsub is a lightweight wrapper of the GCP Pub/Sub functionality.
package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	defaultProjectID       = env.Get("PUBSUB_PROJECT_ID", "", "The project ID of the Pub/Sub.")
	defaultCredentialsFile = env.Get("PUBSUB_CREDENTIALS_FILE", "", "The credentials file of the Pub/Sub project.")
)

// TopicClient is a Pub/Sub client that bound to a topic.
type TopicClient struct {
	topic *pubsub.Topic
}

// NewTopicClient creates a Pub/Sub client that bound to a topic of the given
// project.
func NewTopicClient(projectID, topicID string, opts ...option.ClientOption) (*TopicClient, error) {
	client, err := pubsub.NewClient(context.Background(), projectID, opts...)
	if err != nil {
		return nil, errors.Errorf("create Pub/Sub client: %v", err)
	}
	return &TopicClient{
		topic: client.Topic(topicID),
	}, nil
}

// NewDefaultTopicClient creates a Pub/Sub client that bound to a topic with
// default project ID and credentials file, whose values are read from the
// environment variables `PUBSUB_PROJECT_ID` and `PUBSUB_CREDENTIALS_FILE`
// respectively. It is OK to have empty value for credentials file if the client
// can be authenticated via other means against the target project.
func NewDefaultTopicClient(topicID string) (*TopicClient, error) {
	return NewTopicClient(defaultProjectID, topicID, option.WithCredentialsFile(defaultCredentialsFile))
}

// Ping checks if the connection to the topic is valid.
func (c *TopicClient) Ping(ctx context.Context) error {
	exists, err := c.topic.Exists(ctx)
	if err != nil {
		return err
	} else if !exists {
		return errors.New("topic does not exist")
	}
	return nil
}

// Publish publishes messages and waits for all the results synchronously.
func (c *TopicClient) Publish(ctx context.Context, messages ...[]byte) error {
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

// Stop stops the topic publishing channel. The client should not be used after
// calling Stop.
func (c *TopicClient) Stop() {
	c.topic.Stop()
}
