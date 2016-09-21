package rcache

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"

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

// HACK: Track API cache hits by examining keys on cache requests to see if
// that have the GitHub domain in them. Temporary so we can get better
// visibility into HTTP cache performance.
const githubAPIHost = "api.github.com"

var reposGitHubHTTPCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "repos",
	Name:      "github_api_cache_hit",
	Help:      "Counts cache hits and misses for the github API HTTP cache.",
}, []string{"type"})

// Get implements httpcache.Cache.Get
func (r *Cache) Get(key string) ([]byte, bool) {
	c := pool.Get()
	defer c.Close()

	b, err := redis.Bytes(c.Do("GET", r.rkey(key)))
	if err != nil && err != redis.ErrNil {
		log15.Warn("failed to execute redis command", "cmd", "GET", "error", err)
	}

	// TODO remove this tracking or move it to a more appropriate place.
	if strings.Contains(key, githubAPIHost) {
		go func() {
			if err == nil {
				reposGitHubHTTPCacheCounter.WithLabelValues("hit").Inc()
			} else {
				reposGitHubHTTPCacheCounter.WithLabelValues("miss").Inc()
			}
		}()
	}

	return b, err == nil
}

// Delete implements httpcache.Cache.Set
func (r *Cache) Set(key string, b []byte) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("SETEX", r.rkey(key), r.ttlSeconds, b)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "SETEX", "error", err)
	}
}

// Delete implements httpcache.Cache.Delete
func (r *Cache) Delete(key string) {
	c := pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", r.rkey(key))
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "DEL", "error", err)
	}
}

// rkey generates the actual key we use on redis.
func (r *Cache) rkey(key string) string {
	return fmt.Sprintf("%s:%s:%s", globalPrefix, r.keyPrefix, key)
}

// SetupForTest adjusts the globalPrefix and clears it out. You will have
// conflicts if you do `t.Parallel()`
func SetupForTest(name string) {
	globalPrefix = "__test__" + name
	// Make mutex fails faster
	mutexTries = 1
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("EVAL", `local keys = redis.call('keys', ARGV[1])
if #keys > 0 then
	return redis.call('del', unpack(keys))
else
	return ''
end`, 0, globalPrefix+":*")
	if err != nil {
		log15.Error("Could not clear test prefix", "name", name, "globalPrefix", globalPrefix, "error", err)
	}
}

var (
	pool         *redis.Pool
	globalPrefix string
)

func init() {
	hostname := os.Getenv("SRC_APP_URL")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	globalPrefix = fmt.Sprintf("%s:%s", hostname, dataVersion)

	endpoint := conf.GetenvOrDefault("REDIS_MASTER_ENDPOINT", ":6379")

	pool = &redis.Pool{
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
}
