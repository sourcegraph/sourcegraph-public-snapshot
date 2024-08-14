package rcache

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// dataVersion is used for releases that change type structure for
// data that may already be cached. Increasing this number will
// change the key prefix that is used for all hash keys,
// effectively resetting the cache at the same time the new code
// is deployed.
const dataVersion = "v2"

// Cache implements httpcache.Cache
type Cache struct {
	keyPrefix  string
	ttlSeconds int
	_kv        redispool.KeyValue
}

// New creates a redis backed Cache
func New(kv redispool.KeyValue, keyPrefix string) *Cache {
	return &Cache{
		keyPrefix: keyPrefix,
		_kv:       kv,
	}
}

// NewWithTTL creates a redis backed Cache which expires values after
// ttlSeconds.
func NewWithTTL(kv redispool.KeyValue, keyPrefix string, ttlSeconds int) *Cache {
	return &Cache{
		keyPrefix:  keyPrefix,
		ttlSeconds: ttlSeconds,
		_kv:        kv,
	}
}

func (r *Cache) TTL() time.Duration { return time.Duration(r.ttlSeconds) * time.Second }

// Get implements httpcache.Cache.Get
func (r *Cache) Get(key string) ([]byte, bool) {
	b, err := r.kv().Get(r.rkeyPrefix() + key).Bytes()
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
		err := r.kv().Set(r.rkeyPrefix()+key, b)
		if err != nil {
			log15.Warn("failed to execute redis command", "cmd", "SET", "error", err)
		}
	} else {
		r.SetWithTTL(key, b, r.ttlSeconds)
	}
}

func (r *Cache) SetWithTTL(key string, b []byte, ttl int) {
	if !utf8.Valid([]byte(key)) {
		log15.Error("rcache: keys must be valid utf8", "key", []byte(key))
	}

	err := r.kv().SetEx(r.rkeyPrefix()+key, ttl, b)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "SETEX", "error", err)
	}
}

// SetInt stores an integer value under the specified key in the cache
func (r *Cache) SetInt(key string, value int64) {
	// Convert int to byte slice for storage
	valueStr := strconv.FormatInt(value, 10) // 10 is the base for decimal
	r.Set(key, []byte(valueStr))
}

// GetInt gets an integer value by key. Returns the value and a boolean indicating if the key exists.
func (r *Cache) GetInt64(key string) (int64, bool, error) {
	b, found := r.Get(key)
	if !found {
		return 0, false, nil
	}
	// Correctly convert byte slice to int64
	value, err := strconv.ParseInt(string(b), 10, 64) // 10 is the base, 64 is the bit size
	if err != nil {
		return 0, false, errors.Newf("failed to convert value to int", "value", string(b), "error", err)
	}
	return value, true, nil
}

// IncrByInt64 increments the integer value of a key by the given amount.
// It returns the new value after the increment.
func (r *Cache) IncrByInt64(key string, value int64) (int64, error) {
	newValue, err := r.kv().IncrByInt64(r.rkeyPrefix()+key, value)
	if err != nil {
		return newValue, errors.Newf("failed to execute redis command", "cmd", "INCRBY", "error", err)
	}

	if r.ttlSeconds > 0 {
		// Optionally, set a TTL on the key if ttlSeconds is specified for the cache.
		err = r.kv().Expire(r.rkeyPrefix()+key, r.ttlSeconds)
		if err != nil {
			return newValue, errors.Newf("failed to execute redis command", "cmd", "INCRBY", "error", err)
		}
	}

	return newValue, nil
}

// DecrByInt64 increments the decrements value of a key by the given amount.
// It returns the new value after the increment.
func (r *Cache) DecrByInt64(key string, value int64) (int64, error) {
	newValue, err := r.kv().DecrByInt64(r.rkeyPrefix()+key, value)
	if err != nil {
		return newValue, errors.Newf("failed to execute redis command", "cmd", "DECRBY", "error", err)
	}

	if r.ttlSeconds > 0 {
		// Optionally, set a TTL on the key if ttlSeconds is specified for the cache.
		err = r.kv().Expire(r.rkeyPrefix()+key, r.ttlSeconds)
		if err != nil {
			return newValue, errors.Newf("failed to execute redis command", "cmd", "DECRBY", "error", err)
		}
	}

	return newValue, nil
}

func (r *Cache) Increase(key string) {
	_, err := r.kv().Incr(r.rkeyPrefix() + key)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "INCR", "error", err)
		return
	}

	if r.ttlSeconds <= 0 {
		return
	}

	err = r.kv().Expire(r.rkeyPrefix()+key, r.ttlSeconds)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "EXPIRE", "error", err)
		return
	}
}

func (r *Cache) KeyTTL(key string) (int, bool) {
	ttl, err := r.kv().TTL(r.rkeyPrefix() + key)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "TTL", "error", err)
		return -1, false
	}
	return ttl, ttl >= 0
}

func (r *Cache) ListAllKeys() []string {
	pattern := r.rkeyPrefix() + "*"
	keys, err := r.kv().Keys(pattern)
	if err != nil {
		log15.Warn("failed to execute redis command", "cmd", "KEYS", "pattern", pattern, "error", err)
		return nil
	}
	return keys
}

// FIFOList returns a FIFOList namespaced in r.
func (r *Cache) FIFOList(key string, maxSize int) *FIFOList {
	return NewFIFOList(r.kv(), r.rkeyPrefix()+key, maxSize)
}

// SetHashItem sets a key in a HASH.
// If the HASH does not exist, it is created.
// If the key already exists and is a different type, an error is returned.
// If the hash key does not exist, it is created. If it exists, the value is overwritten.
func (r *Cache) SetHashItem(key string, hashKey string, hashValue string) error {
	return r.kv().HSet(r.rkeyPrefix()+key, hashKey, hashValue)
}

// GetHashItem gets a key in a HASH.
func (r *Cache) GetHashItem(key string, hashKey string) (string, error) {
	return r.kv().HGet(r.rkeyPrefix()+key, hashKey).String()
}

// DeleteHashItem deletes a key in a HASH.
// It returns an integer representing the amount of deleted hash keys:
// If the key exists and the hash key exists, it will return 1.
// If the key exists but the hash key does not, it will return 0.
// If the key does not exist, it will return 0.
func (r *Cache) DeleteHashItem(key string, hashKey string) (int, error) {
	return r.kv().HDel(r.rkeyPrefix()+key, hashKey).Int()
}

// GetHashAll returns the members of the HASH stored at `key`, in no particular order.
func (r *Cache) GetHashAll(key string) (map[string]string, error) {
	return r.kv().HGetAll(r.rkeyPrefix() + key).StringMap()
}

// Delete implements httpcache.Cache.Delete
func (r *Cache) Delete(key string) {
	err := r.kv().Del(r.rkeyPrefix() + key)
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

const testAddr = "127.0.0.1:6379"

// SetupForTest adjusts the globalPrefix and clears it out. You will have
// conflicts if you do `t.Parallel()`. You should always use the returned KeyValue
// in tests. Ultimately, that will help us get rid of the global mock, and the conflicts
// from running tests in parallel.
func SetupForTest(t testing.TB) redispool.KeyValue {
	t.Helper()

	testStore = redispool.NewTestKeyValue()
	t.Cleanup(func() {
		testStore.Pool().Close()
		testStore = nil
	})

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		if err := testStore.Ping(); err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	globalPrefix = "__test__" + t.Name()
	if err := redispool.DeleteAllKeysWithPrefix(testStore, globalPrefix); err != nil {
		log15.Error("Could not clear test prefix", "name", t.Name(), "globalPrefix", globalPrefix, "error", err)
	}

	return testStore
}

var testStore redispool.KeyValue

func (r *Cache) kv() redispool.KeyValue {
	// TODO: We should refactor the SetupForTest method to return a KV, not mock
	// a global thing.
	// That can only work when all tests pass the redis connection directly to the
	// tested methods though.
	if testStore != nil {
		return testStore
	}
	return r._kv
}

var (
	globalPrefix = dataVersion
)
