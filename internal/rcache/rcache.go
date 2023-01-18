package rcache

import (
	"context"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (r *Cache) GetMulti(keys ...string) [][]byte {
	c := poolGet()
	defer c.Close()

	if len(keys) == 0 {
		return nil
	}
	rkeys := make([]any, len(keys))
	for i, key := range keys {
		rkeys[i] = r.rkeyPrefix() + key
	}

	vals, err := redis.Values(c.Do("MGET", rkeys...))
	if err != nil && err != redis.ErrNil {
		log15.Warn("failed to execute redis command", "cmd", "MGET", "error", err)
	}

	strVals := make([][]byte, len(vals))
	for i, val := range vals {
		// MGET returns nil as not found.
		if val == nil {
			continue
		}

		b, err := redis.Bytes(val, nil)
		if err != nil {
			log15.Warn("failed to parse bytes from Redis value", "value", val)
			continue
		}
		strVals[i] = b
	}
	return strVals
}

func (r *Cache) SetMulti(keyvals ...[2]string) {
	c := poolGet()
	defer c.Close()

	if len(keyvals) == 0 {
		return
	}

	for _, kv := range keyvals {
		k, v := kv[0], kv[1]
		if !utf8.Valid([]byte(k)) {
			log15.Error("rcache: keys must be valid utf8", "key", []byte(k))
			continue
		}
		if r.ttlSeconds == 0 {
			if err := c.Send("SET", r.rkeyPrefix()+k, []byte(v)); err != nil {
				log15.Warn("failed to write redis command to client output buffer", "cmd", "SET", "error", err)
			}
		} else {
			if err := c.Send("SETEX", r.rkeyPrefix()+k, r.ttlSeconds, []byte(v)); err != nil {
				log15.Warn("failed to write redis command to client output buffer", "cmd", "SETEX", "error", err)
			}
		}
	}
	if err := c.Flush(); err != nil {
		log15.Warn("failed to flush Redis client", "error", err)
	}
}

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
	c := poolGet()
	defer c.Close()

	if !utf8.Valid([]byte(key)) {
		log15.Error("rcache: keys must be valid utf8", "key", []byte(key))
	}

	if r.ttlSeconds == 0 {
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

// AddToList adds a value to the end of a list.
// If the list does not exist, it is created.
func (r *Cache) AddToList(key string, value string) error {
	c := poolGet()
	defer c.Close()
	_, err := c.Do("RPUSH", r.rkeyPrefix()+key, value)
	return err
}

// GetLastListItems returns the last `count` items in the list.
func (r *Cache) GetLastListItems(key string, count int) ([]string, error) {
	c := poolGet()
	defer c.Close()
	return redis.Strings(c.Do("LRANGE", r.rkeyPrefix()+key, -count, -1))
}

// LTrimList trims the list to the last `count` items.
func (r *Cache) LTrimList(key string, count int) error {
	c := poolGet()
	defer c.Close()
	_, err := c.Do("LTRIM", r.rkeyPrefix()+key, -count, -1)
	return err
}

// DeleteMulti deletes the given keys.
func (r *Cache) DeleteMulti(keys ...string) {
	for _, key := range keys {
		r.Delete(key)
	}
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

// ListKeys lists all keys associated with this cache.
// Use with care if you have long TTLs or no TTL configured.
func (r *Cache) ListKeys(ctx context.Context) (results []string, err error) {
	var c redis.Conn
	c, err = poolGetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get redis conn")
	}
	defer func(c redis.Conn) {
		if tempErr := c.Close(); err == nil {
			err = tempErr
		}
	}(c)

	cursor := 0
	for {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		res, err := redis.Values(
			c.Do("SCAN", cursor,
				"MATCH", r.rkeyPrefix()+"*",
				"COUNT", 100),
		)
		if err != nil {
			return results, errors.Wrap(err, "redis scan")
		}

		cursor, _ = redis.Int(res[0], nil)
		keys, _ := redis.Strings(res[1], nil)
		for i, k := range keys {
			keys[i] = k[len(r.rkeyPrefix()):]
		}

		results = append(results, keys...)

		if cursor == 0 {
			break
		}
	}
	return
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
