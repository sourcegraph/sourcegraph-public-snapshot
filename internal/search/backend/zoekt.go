package backend

import (
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	proto "github.com/sourcegraph/zoekt/grpc/protos/zoekt/webserver/v1"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
)

// ZoektStreamFunc is a convenience function to create a stream receiver from a
// function.
type ZoektStreamFunc func(*zoekt.SearchResult)

func (f ZoektStreamFunc) Send(event *zoekt.SearchResult) {
	f(event)
}

// ZoektDialer is a function that returns a zoekt.Streamer for the given endpoint.
type ZoektDialer func(endpoint string) zoekt.Streamer

// NewCachedZoektDialer wraps a ZoektDialer with caching per endpoint.
func NewCachedZoektDialer(dial ZoektDialer) ZoektDialer {
	d := &cachedZoektDialer{
		streamers: map[string]zoekt.Streamer{},
		dial:      dial,
	}
	return d.Dial
}

type cachedZoektDialer struct {
	mu        sync.RWMutex
	streamers map[string]zoekt.Streamer
	dial      ZoektDialer
}

func (c *cachedZoektDialer) Dial(endpoint string) zoekt.Streamer {
	c.mu.RLock()
	s, ok := c.streamers[endpoint]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		s, ok = c.streamers[endpoint]
		if !ok {
			s = &cachedStreamerCloser{
				cachedZoektDialer: c,
				endpoint:          endpoint,
				Streamer:          c.dial(endpoint),
			}
			c.streamers[endpoint] = s
		}
		c.mu.Unlock()
	}

	return s
}

type cachedStreamerCloser struct {
	*cachedZoektDialer
	endpoint string
	zoekt.Streamer
}

func (c *cachedStreamerCloser) Close() {
	c.mu.Lock()
	delete(c.streamers, c.endpoint)
	c.mu.Unlock()

	c.Streamer.Close()
}

// maxRecvMsgSize is the max message size we can receive from Zoekt without erroring.
// By default, this caps at 4MB, but Zoekt can send payloads significantly larger
// than that depending on the type of search being executed.
// 128MiB is a best guess at reasonable size that will rarely fail.
const maxRecvMsgSize = 128 * 1024 * 1024 // 128MiB

// ZoektDial connects to a Searcher gRPC server at address (host:port).
func ZoektDial(endpoint string) zoekt.Streamer {
	conn, err := defaults.Dial(
		endpoint,
		log.Scoped("zoekt"),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
	)

	// Frontend usually adds a wrapper called automaticRetryClient that will automatically retry requests according
	// to the default.RetryPolicy. For Zoekt, we *do not* use automatic retries, as we want direct control over the
	// error-handling. For example, during search we try to identify if a replica is unreachable, and skip over it
	// and return partial results.
	client := proto.NewWebserverServiceClient(conn)
	return NewMeteredSearcher(endpoint, &zoektGRPCClient{
		endpoint: endpoint,
		client:   client,
		dialErr:  err,
	})
}
