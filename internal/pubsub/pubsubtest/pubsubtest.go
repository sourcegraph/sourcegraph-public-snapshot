package pubsubtest

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
)

var _ pubsub.TopicClient = (*MemoryTopicClient)(nil)

type PublishedMessage struct {
	Data       []byte
	Attributes map[string]string
}

// NewMemoryTopicClient creates a no-op Pub/Sub client does nothing but save
// published messages to an internal slice. Publish is concurrency-safe, but
// reading Messages is not.
func NewMemoryTopicClient() *MemoryTopicClient {
	return &MemoryTopicClient{Messages: []PublishedMessage{}}
}

type MemoryTopicClient struct {
	mux            sync.Mutex
	PrePublishHook func()
	Messages       []PublishedMessage
}

func (c *MemoryTopicClient) Publish(ctx context.Context, messages ...[]byte) error {
	if c.PrePublishHook != nil {
		c.PrePublishHook()
	}
	c.mux.Lock()
	for _, m := range messages {
		c.Messages = append(c.Messages, PublishedMessage{Data: m})
	}
	c.mux.Unlock()
	return nil
}

func (c *MemoryTopicClient) PublishMessage(ctx context.Context, message []byte, attributes map[string]string) error {
	if c.PrePublishHook != nil {
		c.PrePublishHook()
	}
	c.mux.Lock()
	c.Messages = append(c.Messages, PublishedMessage{Data: message, Attributes: attributes})
	c.mux.Unlock()
	return nil
}

func (c *MemoryTopicClient) Ping(context.Context) error { return nil }
func (c *MemoryTopicClient) Stop(context.Context) error { return nil }
