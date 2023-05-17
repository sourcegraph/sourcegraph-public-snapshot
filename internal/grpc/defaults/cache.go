package defaults

import (
	"net/url"
	"strings"
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

	newConn := func(address string) connAndError {
		return newGRPCConnection(address, l)
	}

	return &ConnectionCache{
		connections: ttlcache.New[string, connAndError](newConn, options...),
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
func newGRPCConnection(address string, logger log.Logger) connAndError {
	u, err := parseAddress(address)
	if err != nil {
		return connAndError{
			dialErr: errors.Wrapf(err, "dialing gRPC connection to %q: parsing address %q", address, address),
		}
	}

	gRPCConn, err := Dial(u.Host, logger)
	if err != nil {
		return connAndError{
			dialErr: errors.Wrapf(err, "dialing gRPC connection to %q", address),
		}
	}

	return connAndError{conn: gRPCConn}
}

// parseAddress parses rawAddress into a URL object. It accommodates cases where the rawAddress is a
// simple host:port pair without a URL scheme (e.g., "example.com:8080").
//
// This function aims to provide a flexible way to parse addresses that may or may not strictly adhere to the URL format.
func parseAddress(rawAddress string) (*url.URL, error) {
	addedScheme := false

	// Temporarily prepend "http://" if no scheme is present
	if !strings.Contains(rawAddress, "://") {
		rawAddress = "http://" + rawAddress
		addedScheme = true
	}

	parsedURL, err := url.Parse(rawAddress)
	if err != nil {
		return nil, err
	}

	// If we added the "http://" scheme, remove it from the final URL
	if addedScheme {
		parsedURL.Scheme = ""
	}

	return parsedURL, nil
}

// closeGRPCConnection closes the gRPC connection specified by conn.
func closeGRPCConnection(_ string, conn connAndError) {
	if conn.conn != nil {
		_ = conn.conn.Close()
	}
}

type connAndError struct {
	conn    *grpc.ClientConn
	dialErr error
}
