package backend

import (
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	proto "github.com/sourcegraph/zoekt/grpc/protos/zoekt/webserver/v1"
	"github.com/sourcegraph/zoekt/rpc"
	zoektstream "github.com/sourcegraph/zoekt/stream"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// We don't use the normal factory for internal requests because we disable
// retries. Currently our retry framework copies the full body on every
// request, this is prohibitive when zoekt generates a large query.
//
// Once our retry framework supports the use of Request.GetBody we can switch
// back to the normal internal request factory.
var zoektHTTPClient, _ = httpcli.NewFactory(
	httpcli.NewMiddleware(
		httpcli.ContextErrorMiddleware,
	),
	httpcli.NewMaxIdleConnsPerHostOpt(500),
	// This will also generate a metric named "src_zoekt_webserver_requests_total".
	httpcli.MeteredTransportOpt("zoekt_webserver"),
	httpcli.TracedTransportOpt,
).Client()

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

// ZoektDial connects to a Searcher HTTP RPC server at address (host:port).
func ZoektDial(endpoint string) zoekt.Streamer {
	return &switchableZoektGRPCClient{
		httpClient: ZoektDialHTTP(endpoint),
		grpcClient: ZoektDialGRPC(endpoint),
	}
}

// ZoektDialHTTP connects to a Searcher HTTP RPC server at address (host:port).
func ZoektDialHTTP(endpoint string) zoekt.Streamer {
	client := rpc.Client(endpoint)
	streamClient := zoektstream.NewClient("http://"+endpoint, zoektHTTPClient).WithSearcher(client)
	return NewMeteredSearcher(endpoint, streamClient)
}

// maxRecvMsgSize is the max message size we can receive from Zoekt without erroring.
// By default, this caps at 4MB, but Zoekt can send payloads significantly larger
// than that depending on the type of search being executed.
// 128MiB is a best guess at reasonable size that will rarely fail.
const maxRecvMsgSize = 128 * 1024 * 1024 // 128MiB

// ZoektDialGRPC connects to a Searcher gRPC server at address (host:port).
func ZoektDialGRPC(endpoint string) zoekt.Streamer {
	conn, err := defaults.Dial(
		endpoint,
		log.Scoped("zoekt"),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
	)
	return NewMeteredSearcher(endpoint, &zoektGRPCClient{
		endpoint: endpoint,
		client:   proto.NewWebserverServiceClient(conn),
		dialErr:  err,
	})
}
