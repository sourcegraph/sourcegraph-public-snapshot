package pubsubtest

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
)

var _ pubsub.TopicClient = (*MemoryTopicClient)(nil)

// NewMemoryTopicClient creates a no-op Pub/Sub client does nothing but save
// published messages to an internal slice. Publish is concurrency-safe, but
// reading Messages is not.
func NewMemoryTopicClient() *MemoryTopicClient {
	return &MemoryTopicClient{Messages: [][]byte{}}
}

type MemoryTopicClient struct {
	mux            sync.Mutex
	PrePublishHook func()
	Messages       [][]byte
}

func (c *MemoryTopicClient) Publish(ctx context.Context, messages ...[]byte) error {
	if c.PrePublishHook != nil {
		c.PrePublishHook()
	}
	c.mux.Lock()
	c.Messages = append(c.Messages, messages...)
	c.mux.Unlock()
	return nil
}

func (c *MemoryTopicClient) Ping(context.Context) error { return nil }
func (c *MemoryTopicClient) Stop()                      {}
