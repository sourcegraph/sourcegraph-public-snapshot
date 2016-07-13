package rcache

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

const (
	maxClients = 32

	// dataVersion is used for releases that change type struture for
	// data that may already be cached. Increasing this number will
	// change the key prefix that is used for all hash keys,
	// effectively resetting the cache at the same time the new code
	// is deployed.
	dataVersion = "v1"
)

var (
	connPool_    *pool.Pool
	connPoolMu   sync.Mutex
	globalPrefix string
)

// redisPool creates the Redis connection pool if it isn't already
// open and returns it. Subsequent calls return the same pool.
func redisPool() (*pool.Pool, error) {
	connPoolMu.Lock()
	defer connPoolMu.Unlock()

	if connPool_ != nil {
		return connPool_, nil
	}

	hostname := os.Getenv("SRC_APP_URL")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	globalPrefix = fmt.Sprintf("%s:%s", hostname, dataVersion)

	endpoint := conf.GetenvOrDefault("REDIS_MASTER_ENDPOINT", ":6379")

	p, err := pool.New("tcp", endpoint, maxClients)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to Redis server at %s: %s", endpoint, err)
	}
	connPool_ = p

	return connPool_, nil
}

// getConn returns a redis client from the pool. When you are done you must
// call the cleanup function to return the connection to the pool.
func getConn() (*redis.Client, func(), error) {
	connPool, err := redisPool()
	if err != nil {
		return nil, nil, err
	}

	conn, err := connPool.Get()
	if err != nil {
		return nil, nil, err
	}
	return conn, func() { connPool.Put(conn) }, nil
}

// Redis exposes methods for speaking to the configured redis server. It
// serves two main functions: Ensure we namespace keys and correct usage of
// the connection pool.
type Redis struct {
	// keyPrefix is the prefix that is prepended to each key stored in
	// Redis by this cache.
	keyPrefix string
}

var ErrNotFound = errors.New("Redis key not found")

// Get is the Redis GET command. error is ErrNotFound if missing.
func (r *Redis) Get(key string) ([]byte, error) {
	conn, cleanup, err := getConn()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := conn.Cmd("GET", r.rkey(key))
	if resp.IsType(redis.Nil) {
		return nil, ErrNotFound
	}
	if resp.Err != nil {
		return nil, fmt.Errorf("Redis.Get error: %s", resp.Err)
	}

	return resp.Bytes()
}

// Del is the Redis DEL command.
func (r *Redis) Del(key string) error {
	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	resp := conn.Cmd("DEL", r.rkey(key))
	if resp.Err != nil {
		return fmt.Errorf("Redis.Del error: %s", resp.Err)
	}
	return nil
}

// Set is the Redis SET command.
func (r *Redis) Set(key string, val []byte) error {
	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	resp := conn.Cmd("SET", r.rkey(key), val)
	if resp.Err != nil {
		return fmt.Errorf("Redis.Add error: %s", resp.Err)
	}
	return nil
}

// SetEx is the Redis SETEX command.
func (r *Redis) SetEx(key string, val []byte, ttlSeconds int) error {
	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	resp := conn.Cmd("SETEX", r.rkey(key), ttlSeconds, val)
	if resp.Err != nil {
		return fmt.Errorf("Redis.Add error: %s", resp.Err)
	}
	return nil
}

// rkey generates the actual key we use on redis.
func (r *Redis) rkey(key string) string {
	return fmt.Sprintf("%s:%s:%s", globalPrefix, r.keyPrefix, key)
}

// Cache implements httpcache.Cache
type Cache struct {
	r          *Redis
	ttlSeconds int
}

// New creates a redis backed Cache
func New(keyPrefix string, ttlSeconds int) *Cache {
	return &Cache{
		r:          &Redis{keyPrefix: keyPrefix},
		ttlSeconds: ttlSeconds,
	}
}

// Get implements httpcache.Cache.Get
func (r *Cache) Get(key string) ([]byte, bool) {
	b, err := r.r.Get(key)
	return b, err == nil
}

// Delete implements httpcache.Cache.Set
func (r *Cache) Set(key string, responseBytes []byte) {
	_ = r.r.SetEx(key, responseBytes, r.ttlSeconds)
}

// Delete implements httpcache.Cache.Delete
func (r *Cache) Delete(key string) {
	_ = r.r.Del(key)
}

// ClearAllForTest clears all of the entries with a given prefix. This
// is an O(n) operation and should only be used in tests.
func ClearAllForTest(prefix string) error {
	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	resp := conn.Cmd("EVAL", `local keys = redis.call('keys', ARGV[1])
if #keys > 0 then
	return redis.call('del', unpack(keys))
else
	return ''
end`, 0, fmt.Sprintf("%s:*", fmt.Sprintf("%s:%s", globalPrefix, prefix)))
	if resp.Err != nil {
		return fmt.Errorf("error clearing Redis test data: %s", resp.Err)
	}
	return nil
}
