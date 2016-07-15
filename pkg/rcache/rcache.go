package rcache

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
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

// Cache implements httpcache.Cache
type Cache struct {
	keyPrefix  string
	ttlSeconds int
}

// New creates a redis backed Cache
func New(keyPrefix string, ttlSeconds int) *Cache {
	return &Cache{
		keyPrefix:  keyPrefix,
		ttlSeconds: ttlSeconds,
	}
}

// Get implements httpcache.Cache.Get
func (r *Cache) Get(key string) ([]byte, bool) {
	resp := cmd("GET", r.rkey(key))
	if resp == nil {
		return nil, false
	}

	b, err := redis.Bytes(resp, nil)
	if err != nil {
		return nil, false
	}
	return b, true
}

// Delete implements httpcache.Cache.Set
func (r *Cache) Set(key string, b []byte) {
	_ = cmd("SETEX", r.rkey(key), r.ttlSeconds, b)
}

// Delete implements httpcache.Cache.Delete
func (r *Cache) Delete(key string) {
	_ = cmd("DEL", r.rkey(key))
}

// rkey generates the actual key we use on redis.
func (r *Cache) rkey(key string) string {
	return fmt.Sprintf("%s:%s:%s", globalPrefix, r.keyPrefix, key)
}

// ClearAllForTest clears all of the entries with a given prefix. This
// is an O(n) operation and should only be used in tests.
func ClearAllForTest(prefix string) error {
	resp := cmd("EVAL", `local keys = redis.call('keys', ARGV[1])
if #keys > 0 then
	return redis.call('del', unpack(keys))
else
	return ''
end`, 0, fmt.Sprintf("%s:*", fmt.Sprintf("%s:%s", globalPrefix, prefix)))
	if err, ok := resp.(error); ok {
		return fmt.Errorf("error clearing Redis test data: %s", err)
	}
	return nil
}

var (
	connPool_    *redis.Pool
	connPoolMu   sync.Mutex
	globalPrefix string
)

// redisPool creates the Redis connection pool if it isn't already
// open and returns it. Subsequent calls return the same pool.
func redisPool() (*redis.Pool, error) {
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

	connPool_ := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", endpoint)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return connPool_, nil
}

// cmd is a helper around redis.(*Client).Cmd. If any error happens (including
// resp.Err) cmd will log it and return nil.
func cmd(cmd string, args ...interface{}) interface{} {
	connPool, err := redisPool()
	if err != nil {
		log15.Warn("failed to connect to redis pool", "error", err)
		return nil
	}
	conn := connPool.Get()
	defer conn.Close()

	r, err := conn.Do(cmd, args...)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", cmd, "error", err)
		return nil
	}
	return r
}
