package rcache

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

// dataVersion is used for releases that change type structure for
// data that may already be cached. Increasing this number will
// change the key prefix that is used for all hash keys,
// effectively resetting the cache at the same time the new code
// is deployed.
const dataVersion = "v2"
const dataVersionToDelete = "v1"

// DeleteOldCacheData deletes the rcache data in the given Redis instance
// that's prefixed with dataVersionToDelete
func DeleteOldCacheData(c redis.Conn) error {
	return deleteAllKeysWithPrefix(c, dataVersionToDelete)
}

// Cache implements httpcache.Cache
type Cache struct {
	keyPrefix  string
	ttlSeconds int
}

// New creates a redis backed Cache
func New(keyPrefix string) *Cache {
	return &Cache{
		keyPrefix: keyPrefix,
	}
}

// NewWithTTL creates a redis backed Cache which expires values after
// ttlSeconds.
func NewWithTTL(keyPrefix string, ttlSeconds int) *Cache {
	return &Cache{
		keyPrefix:  keyPrefix,
		ttlSeconds: ttlSeconds,
	}
}

func (r *Cache) TTL() time.Duration { return time.Duration(r.ttlSeconds) * time.Second }

// Get implements httpcache.Cache.Get
func (r *Cache) Get(key string) ([]byte, bool) {
	c := poolGet()
	defer c.Close()

	b, err := redis.Bytes(c.Do("GET", r.rkeyPrefix()+key))
	if err != nil && err != redis.ErrNil {
		log15.Warn("failed to execute redis command", "cmd", "GET", "error", err)
	}

	return b, err == nil
}

// Set implements httpcache.Cache.Set
func (r *Cache) Set(key string, b []byte) {
	if !utf8.Valid([]byte(key)) {
		log15.Error("rcache: keys must be valid utf8", "key", []byte(key))
	}

	if r.ttlSeconds == 0 {
		c := poolGet()
		defer c.Close()

		_, err := c.Do("SET", r.rkeyPrefix()+key, b)
		if err != nil {
			log15.Warn("failed to execute redis command", "cmd", "SET", "error", err)
		}
	} else {
		r.SetWithTTL(key, b, r.ttlSeconds)
	}
}

func (r *Cache) SetWithTTL(key string, b []byte, ttl int) {
	c := poolGet()
	defer c.Close()

	if !utf8.Valid([]byte(key)) {
		log15.Error("rcache: keys must be valid utf8", "key", []byte(key))
	}

	_, err := c.Do("SETEX", r.rkeyPrefix()+key, ttl, b)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "SETEX", "error", err)
	}
}

func (r *Cache) Increase(key string) {
	c := poolGet()
	defer func() { _ = c.Close() }()

	_, err := c.Do("INCR", r.rkeyPrefix()+key)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "INCR", "error", err)
		return
	}

	if r.ttlSeconds <= 0 {
		return
	}

	_, err = c.Do("EXPIRE", r.rkeyPrefix()+key, r.ttlSeconds)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "EXPIRE", "error", err)
		return
	}
}

func (r *Cache) KeyTTL(key string) (int, bool) {
	c := poolGet()
	defer func() { _ = c.Close() }()

	ttl, err := redis.Int(c.Do("TTL", r.rkeyPrefix()+key))
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "TTL", "error", err)
		return -1, false
	}
	return ttl, ttl >= 0
}

// FIFOList returns a FIFOList namespaced in r.
func (r *Cache) FIFOList(key string, maxSize int) *FIFOList {
	return NewFIFOList(r.rkeyPrefix()+key, maxSize)
}

// SetHashItem sets a key in a HASH.
// If the HASH does not exist, it is created.
// If the key already exists and is a different type, an error is returned.
// If the hash key does not exist, it is created. If it exists, the value is overwritten.
func (r *Cache) SetHashItem(key string, hashKey string, hashValue string) error {
	c := poolGet()
	defer c.Close()
	_, err := c.Do("HSET", r.rkeyPrefix()+key, hashKey, hashValue)
	return err
}

// GetHashItem gets a key in a HASH.
func (r *Cache) GetHashItem(key string, hashKey string) (string, error) {
	c := poolGet()
	defer c.Close()
	return redis.String(c.Do("HGET", r.rkeyPrefix()+key, hashKey))
}

// GetHashAll returns the members of the HASH stored at `key`, in no particular order.
func (r *Cache) GetHashAll(key string) (map[string]string, error) {
	c := poolGet()
	defer c.Close()
	return redis.StringMap(c.Do("HGETALL", r.rkeyPrefix()+key))
}

// Delete implements httpcache.Cache.Delete
func (r *Cache) Delete(key string) {
	c := poolGet()
	defer func(c redis.Conn) {
		_ = c.Close()
	}(c)

	_, err := c.Do("DEL", r.rkeyPrefix()+key)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "DEL", "error", err)
	}
}

// rkeyPrefix generates the actual key prefix we use on redis.
func (r *Cache) rkeyPrefix() string {
	return fmt.Sprintf("%s:%s:", globalPrefix, r.keyPrefix)
}

// TB is a subset of testing.TB
type TB interface {
	Name() string
	Skip(args ...any)
	Helper()
}

// SetupForTest adjusts the globalPrefix and clears it out. You will have
// conflicts if you do `t.Parallel()`
func SetupForTest(t TB) {
	t.Helper()

	pool = redispool.RedisKeyValue(&redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	})

	globalPrefix = "__test__" + t.Name()
	c := poolGet()
	defer c.Close()

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	err := deleteAllKeysWithPrefix(c, globalPrefix)
	if err != nil {
		log15.Error("Could not clear test prefix", "name", t.Name(), "globalPrefix", globalPrefix, "error", err)
	}
}

// The number of keys to delete per batch.
// The maximum number of keys that can be unpacked
// is determined by the Lua config LUAI_MAXCSTACK
// which is 8000 by default.
// See https://www.lua.org/source/5.1/luaconf.h.html
var deleteBatchSize = 5000

func deleteAllKeysWithPrefix(c redis.Conn, prefix string) error {
	const script = `
redis.replicate_commands()
local cursor = '0'
local prefix = ARGV[1]
local batchSize = ARGV[2]
local result = ''
repeat
	local keys = redis.call('SCAN', cursor, 'MATCH', prefix, 'COUNT', batchSize)
	if #keys[2] > 0
	then
		result = redis.call('DEL', unpack(keys[2]))
	end

	cursor = keys[1]
until cursor == '0'
return result
`

	_, err := c.Do("EVAL", script, 0, prefix+":*", deleteBatchSize)
	return err
}

var (
	pool         = redispool.Cache
	globalPrefix = dataVersion
)
