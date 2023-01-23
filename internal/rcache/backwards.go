package rcache

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/gomodule/redigo/redis"
)

// poolGet is temporary wrapper around getting a raw redis connection. It just
// fails if redis is disabled. We intend to remove this wrapper in the future
// and instead fallback to in-memory options for the Sourcegraph App.
func poolGet() redis.Conn {
	pool, ok := pool.Pool()
	if !ok {
		return errorConn{err: errRedisDisable}
	}

	return pool.Get()
}

func poolGetContext(ctx context.Context) (redis.Conn, error) {
	pool, ok := pool.Pool()
	if !ok {
		return errorConn{err: errRedisDisable}, nil
	}

	return pool.GetContext(ctx)
}

var errRedisDisable = errors.New("redis is disabled")

// copy pasta from redigo/redis
type errorConn struct{ err error }

func (ec errorConn) Do(string, ...interface{}) (interface{}, error) { return nil, ec.err }
func (ec errorConn) DoWithTimeout(time.Duration, string, ...interface{}) (interface{}, error) {
	return nil, ec.err
}
func (ec errorConn) Send(string, ...interface{}) error                     { return ec.err }
func (ec errorConn) Err() error                                            { return ec.err }
func (ec errorConn) Close() error                                          { return nil }
func (ec errorConn) Flush() error                                          { return ec.err }
func (ec errorConn) Receive() (interface{}, error)                         { return nil, ec.err }
func (ec errorConn) ReceiveWithTimeout(time.Duration) (interface{}, error) { return nil, ec.err }
