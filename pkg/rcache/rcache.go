package rcache

import (
	"encoding/json"
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

// Redis is a cache implemented on top of a Redis client. It is
// designed to mimick the API of cache.Cache to make it easy to switch
// instances of cache.Cache to Redis.
type Redis struct {
	// keyPrefix is the prefix that is prepended to each key stored in
	// Redis by this cache.
	keyPrefix string
}

func New(keyPrefix string) *Redis {
	return &Redis{
		keyPrefix: keyPrefix,
	}
}

var ErrNotFound = errors.New("Redis key not found")

// Get fetches the cached value for the given key into the
// destination. If the key does not exist, it will return ErrNotFound.
func (r *Redis) Get(key string, dst interface{}) error {
	rkey := fmt.Sprintf("%s:%s:%s", globalPrefix, r.keyPrefix, key)

	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	resp := conn.Cmd("GET", rkey)
	if resp.IsType(redis.Nil) {
		return ErrNotFound
	}
	if resp.Err != nil {
		return fmt.Errorf("Redis.Get error: %s", resp.Err)
	}

	b, err := resp.Bytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return err
	}
	return nil
}

// Add adds a value to the Redis-backed cache with the specified key.
// If ttlSeconds =< 0, then a TTL will not be set.
func (r *Redis) Add(key string, val interface{}, ttlSeconds int) error {
	rkey := fmt.Sprintf("%s:%s:%s", globalPrefix, r.keyPrefix, key)

	conn, cleanup, err := getConn()
	if err != nil {
		return err
	}
	defer cleanup()

	vjson, err := json.Marshal(val)
	if err != nil {
		return err
	}

	if ttlSeconds <= 0 {
		resp := conn.Cmd("SET", rkey, vjson)
		if resp.Err != nil {
			return fmt.Errorf("Redis.Add error: %s", resp.Err)
		}
	} else {
		resp := conn.Cmd("SETEX", rkey, ttlSeconds, vjson)
		if resp.Err != nil {
			return fmt.Errorf("Redis.Add error: %s", resp.Err)
		}
	}
	return nil
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
