package defaults

import (
	"net/url"
	"time"

	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/ttlcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ConnectionCache is a cache of gRPC connections. It is safe for concurrent use.
type ConnectionCache struct {
	connections *ttlcache.Cache[string, connAndError]

	started chan struct{}
}

// NewConnectionCache creates a new ConnectionCache. The cache must be started with Start before
// it can be used.
//
// This cache will close gRPC connections after 10 minutes of inactivity.
func NewConnectionCache() *ConnectionCache {
	options := []ttlcache.Option[string, connAndError]{
		ttlcache.WithExpirationFunc[string, connAndError](closeGRPCConnection),
		ttlcache.WithReapInterval[string, connAndError](1 * time.Minute),
		ttlcache.WithExpirationTime[string, connAndError](10 * time.Minute),
	}

	return &ConnectionCache{
		connections: ttlcache.New[string, connAndError](newGRPCConnection, options...),
		started:     make(chan struct{}),
	}
}

// Start starts the routines that reap expired connections. This must be called before
// GetConnection can be used.
func (c *ConnectionCache) Start() {
	select {
	case <-c.started:
		return
	default:
		c.connections.StartReaper()
		close(c.started)
	}
}

// Shutdown closes the cache. This must be called when the cache is no longer needed.
func (c *ConnectionCache) Shutdown() {
	c.connections.Shutdown()
}

// GetConnection returns a gRPC connection to the given address. If the connection is not in the
// cache, a new connection will be created.
//
// GetConnection will block until the cache is started with Start.
func (c *ConnectionCache) GetConnection(address string) (*grpc.ClientConn, error) {
	<-c.started

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
