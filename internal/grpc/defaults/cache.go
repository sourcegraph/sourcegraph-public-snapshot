package defaults

import (
	"net/url"
	"sync"
	"time"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/ttlcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ConnectionCache is a cache of gRPC connections. It is safe for concurrent use.
//
// When the cache is no longer needed, Shutdown should be called to release resources associated
// with the cache.
type ConnectionCache struct {
	connections *ttlcache.Cache[string, connAndError]

	startOnce sync.Once
}

// NewConnectionCache creates a new ConnectionCache. When the cache is no longer needed, Shutdown
// should be called to release resources associated with the cache.
//
// This cache will close gRPC connections after 10 minutes of inactivity.
func NewConnectionCache(l log.Logger) *ConnectionCache {
	options := []ttlcache.Option[string, connAndError]{
		ttlcache.WithExpirationFunc[string, connAndError](closeGRPCConnection),

		ttlcache.WithReapInterval[string, connAndError](1 * time.Minute),
		ttlcache.WithTTL[string, connAndError](10 * time.Minute),

		ttlcache.WithLogger[string, connAndError](l),

		// 1000 connections is a lot. If we ever hit this, we should probably
		// warn so we can investigate.
		ttlcache.WithSizeWarningThreshold[string, connAndError](1000),
	}

	return &ConnectionCache{
		connections: ttlcache.New[string, connAndError](newGRPCConnection, options...),
	}
}

// ensureStarted starts the routines that reap expired connections.
func (c *ConnectionCache) ensureStarted() {
	c.startOnce.Do(c.connections.StartReaper)
}

// Shutdown tears down the background goroutines that maintain the cache.
// This should be called when the cache is no longer needed.
func (c *ConnectionCache) Shutdown() {
	c.connections.Shutdown()
}

// GetConnection returns a gRPC connection to the given address. If the connection is not in the
// cache, a new connection will be created.
func (c *ConnectionCache) GetConnection(address string) (*grpc.ClientConn, error) {
	c.ensureStarted()

	ce := c.connections.Get(address)
	return ce.conn, ce.dialErr
}

// newGRPCConnection creates a new gRPC connection to the given address, or returns an error if
// the connection could not be created.
func newGRPCConnection(address string) connAndError {
	u, err := url.Parse(address)
	if err != nil {
		return connAndError{
			dialErr: errors.Wrapf(err, "parsing address %q", address),
		}
	}

	gRPCConn, err := Dial(u.Host)
	if err != nil {
		return connAndError{
			dialErr: errors.Wrapf(err, "dialing gRPC connection to %q", address),
		}
	}

	return connAndError{conn: gRPCConn}
}

// closeGRPCConnection closes the gRPC connection specified by conn.
func closeGRPCConnection(_ string, conn connAndError) {
	if conn.dialErr != nil {
		_ = conn.conn.Close()
	}
}

type connAndError struct {
	conn    *grpc.ClientConn
	dialErr error
}
