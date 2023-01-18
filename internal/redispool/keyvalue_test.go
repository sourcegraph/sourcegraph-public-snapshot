package redispool_test

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRedisKeyValue(t *testing.T) {
	testKeyValue(t, redisKeyValueForTest(t))
}

func testKeyValue(t *testing.T, kv redispool.KeyValue) {
	// assertWorks is a weird name, but it makes the function name align with
	// assertEqual.
	assertWorks := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal("unexpected error: ", err)
		}
	}
	// redispool.Value helpers to make test readable
	assertEqual := func(got redispool.Value, want any) {
		t.Helper()
		switch wantV := want.(type) {
		case bool:
			gotV, err := got.Bool()
			assertWorks(err)
			if gotV != wantV {
				t.Fatalf("got %v, wanted %v", gotV, wantV)
			}
		case []byte:
			gotV, err := got.Bytes()
			assertWorks(err)
			if !reflect.DeepEqual(gotV, wantV) {
				t.Fatalf("got %q, wanted %q", gotV, wantV)
			}
		case string:
			gotV, err := got.String()
			assertWorks(err)
			if gotV != wantV {
				t.Fatalf("got %q, wanted %q", gotV, wantV)
			}
		case nil:
			_, err := got.String()
			if err != redis.ErrNil {
				t.Fatalf("%v is not nil", got)
			}
		case error:
			gotV, err := got.String()
			if err == nil {
				t.Fatalf("want error, got %q", gotV)
			}
			if !strings.Contains(err.Error(), wantV.Error()) {
				t.Fatalf("got error %v, wanted error %v", err, wantV)
			}
		default:
			t.Fatalf("unsupported want type for %q: %T", want, want)
		}
	}

	// Redis returns nil on unset values
	assertEqual(kv.Get("hi"), redis.ErrNil)

	// Simple get followed by set. Redigo autocasts, ensure we keep that
	// behaviour.
	assertWorks(kv.Set("simple", "1"))
	assertEqual(kv.Get("simple"), "1")
	assertEqual(kv.Get("simple"), true)
	assertEqual(kv.Get("simple"), []byte("1"))

	// GetSet on existing value
	assertEqual(kv.GetSet("simple", "2"), "1")
	assertEqual(kv.GetSet("simple", "3"), "2")
	assertEqual(kv.Get("simple"), "3")

	// GetSet on nil value
	assertEqual(kv.GetSet("missing", "found"), redis.ErrNil)
	assertEqual(kv.Get("missing"), "found")
	assertWorks(kv.Del("missing"))
	assertEqual(kv.Get("missing"), redis.ErrNil)

	// Ensure we can handle funky bytes
	assertWorks(kv.Set("funky", []byte{0, 10, 100, 255}))
	assertEqual(kv.Get("funky"), []byte{0, 10, 100, 255})

	// Ensure we fail hashes when used against non hashes.
	assertEqual(kv.HGet("simple", "field"), errors.New("WRONGTYPE"))
	if err := kv.HSet("simple", "field", "value"); !strings.Contains(err.Error(), "WRONGTYPE") {
		t.Fatalf("expected wrongtype error, got %v", err)
	}

	// Pretty much copy-pasta above tests but on a hash

	// Redis returns nil on unset hashes
	assertEqual(kv.HGet("hash", "hi"), redis.ErrNil)

	// Simple hget followed by hset. Redigo autocasts, ensure we keep that
	// behaviour.
	assertWorks(kv.HSet("hash", "simple", "1"))
	assertEqual(kv.HGet("hash", "simple"), "1")
	assertEqual(kv.HGet("hash", "simple"), true)
	assertEqual(kv.HGet("hash", "simple"), []byte("1"))

	// Redis returns nil on unset fields
	assertEqual(kv.HGet("hash", "hi"), redis.ErrNil)

	// Ensure we can handle funky bytes
	assertWorks(kv.HSet("hash", "funky", []byte{0, 10, 100, 255}))
	assertEqual(kv.HGet("hash", "funky"), []byte{0, 10, 100, 255})

	// We intentionally do not test EXPIRE since I don't like sleeps in tests.
}

// Mostly copy-pasta from rache. Will clean up later as the relationship
// between the two packages becomes cleaner.
func redisKeyValueForTest(t *testing.T) redispool.KeyValue {
	t.Helper()

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	prefix := "__test__" + t.Name()
	c := pool.Get()
	defer c.Close()

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	if err := deleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	kv := redispool.RedisKeyValue(pool).(interface {
		WithPrefix(string) redispool.KeyValue
	})
	return kv.WithPrefix(prefix)
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
