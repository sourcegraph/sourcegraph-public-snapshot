package redispool_test

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRedisKeyValue(t *testing.T) {
	testKeyValue(t, redisKeyValueForTest(t))
}

func testKeyValue(t *testing.T, kv redispool.KeyValue) {
	t.Parallel()

	errWrongType := errors.New("WRONGTYPE")

	// "strings" is the name of the classic group of commands in redis (get, set, ttl, etc). We call it classic since that is less confusing.
	t.Run("classic", func(t *testing.T) {
		t.Parallel()

		require := require{TB: t}

		// Redis returns nil on unset values
		require.Equal(kv.Get("hi"), redis.ErrNil)

		// Simple get followed by set. Redigo autocasts, ensure we keep that
		// behaviour.
		require.Works(kv.Set("simple", "1"))
		require.Equal(kv.Get("simple"), "1")
		require.Equal(kv.Get("simple"), 1)
		require.Equal(kv.Get("simple"), true)
		require.Equal(kv.Get("simple"), []byte("1"))

		// Set when not exists
		set, err := kv.SetNx("setnx", "2")
		require.Works(err)
		assert.True(t, set)
		set, err = kv.SetNx("setnx", "3")
		require.Works(err)
		assert.False(t, set)
		require.Equal(kv.Get("setnx"), "2")

		// GetSet on existing value
		require.Equal(kv.GetSet("simple", "2"), "1")
		require.Equal(kv.GetSet("simple", "3"), "2")
		require.Equal(kv.Get("simple"), "3")

		// GetSet on nil value
		require.Equal(kv.GetSet("missing", "found"), redis.ErrNil)
		require.Equal(kv.Get("missing"), "found")
		require.Works(kv.Del("missing"))
		require.Equal(kv.Get("missing"), redis.ErrNil)

		// Ensure we can handle funky bytes
		require.Works(kv.Set("funky", []byte{0, 10, 100, 255}))
		require.Equal(kv.Get("funky"), []byte{0, 10, 100, 255})

		// Incr
		require.Works(kv.Set("incr-set", 5))
		_, err = kv.Incr("incr-set")
		require.Works(err)
		_, err = kv.Incr("incr-unset")
		require.Works(err)
		require.Equal(kv.Get("incr-set"), 6)
		require.Equal(kv.Get("incr-unset"), 1)

		// Incrby
		require.Works(kv.Set("incrby-set", 5))
		_, err = kv.Incrby("incrby-set", 2)
		require.Works(err)
		_, err = kv.Incrby("incrby-unset", 2)
		require.Works(err)
		require.Equal(kv.Get("incrby-set"), 7)
		require.Equal(kv.Get("incrby-unset"), 2)
	})

	t.Run("hash", func(t *testing.T) {
		t.Parallel()

		require := require{TB: t}

		// Pretty much copy-pasta above tests but on a hash

		// Redis returns nil on unset hashes
		require.Equal(kv.HGet("hash", "hi"), redis.ErrNil)

		// Simple hget followed by hset. Redigo autocasts, ensure we keep that
		// behaviour.
		require.Works(kv.HSet("hash", "simple", "1"))
		require.Equal(kv.HGet("hash", "simple"), "1")
		require.Equal(kv.HGet("hash", "simple"), true)
		require.Equal(kv.HGet("hash", "simple"), []byte("1"))

		// hgetall
		require.Works(kv.HSet("hash", "horse", "graph"))
		require.AllEqual(kv.HGetAll("hash"), map[string]string{
			"simple": "1",
			"horse":  "graph",
		})

		// hdel and ensure it no longer exists
		require.Equal(kv.HDel("hash", "horse"), 1)
		require.Equal(kv.HGet("hash", "horse"), redis.ErrNil)
		// Nonexistent key returns 0
		require.Equal(kv.HGet("doesnotexist", "neitherdoesthis"), redis.ErrNil)
		require.Equal(kv.HDel("doesnotexist", "neitherdoesthis"), 0)
		// Existing key but nonexistent field returns 0
		require.Equal(kv.HGet("hash", "doesnotexist"), redis.ErrNil)
		require.Equal(kv.HDel("hash", "doesnotexist"), 0)

		// Redis returns nil on unset fields
		require.Equal(kv.HGet("hash", "hi"), redis.ErrNil)

		// Ensure we can handle funky bytes
		require.Works(kv.HSet("hash", "funky", []byte{0, 10, 100, 255}))
		require.Equal(kv.HGet("hash", "funky"), []byte{0, 10, 100, 255})
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()

		require := require{TB: t}

		// Redis behaviour on unset lists
		require.ListLen(kv, "list-unset-0", 0)
		require.AllEqual(kv.LRange("list-unset-1", 0, 10), bytes())
		require.Works(kv.LTrim("list-unset-2", 0, 10))

		require.Works(kv.LPush("list", "4"))
		require.Works(kv.LPush("list", "3"))
		require.Works(kv.LPush("list", "2"))
		require.Works(kv.LPush("list", "1"))
		require.Works(kv.LPush("list", "0"))

		// Different ways we get the full list back
		require.AllEqual(kv.LRange("list", 0, 10), []string{"0", "1", "2", "3", "4"})
		require.AllEqual(kv.LRange("list", 0, 10), bytes("0", "1", "2", "3", "4"))
		require.AllEqual(kv.LRange("list", 0, -1), bytes("0", "1", "2", "3", "4"))
		require.AllEqual(kv.LRange("list", -5, -1), bytes("0", "1", "2", "3", "4"))
		require.AllEqual(kv.LRange("list", 0, 4), bytes("0", "1", "2", "3", "4"))

		// If stop < start we return nothing
		require.AllEqual(kv.LRange("list", -1, 0), bytes())

		// Subsets
		require.AllEqual(kv.LRange("list", 1, 3), bytes("1", "2", "3"))
		require.AllEqual(kv.LRange("list", 1, -2), bytes("1", "2", "3"))
		require.AllEqual(kv.LRange("list", -4, 3), bytes("1", "2", "3"))
		require.AllEqual(kv.LRange("list", -4, -2), bytes("1", "2", "3"))

		// Trim noop
		require.Works(kv.LTrim("list", 0, 10))
		require.AllEqual(kv.LRange("list", 0, 4), bytes("0", "1", "2", "3", "4"))

		// Trim popback
		require.Works(kv.LTrim("list", 0, -2))
		require.AllEqual(kv.LRange("list", 0, 4), bytes("0", "1", "2", "3"))
		require.ListLen(kv, "list", 4)

		// Trim popfront
		require.Works(kv.LTrim("list", 1, 10))
		require.AllEqual(kv.LRange("list", 0, 4), bytes("1", "2", "3"))
		require.ListLen(kv, "list", 3)

		// Trim all
		require.Works(kv.LTrim("list", -1, -2))
		require.AllEqual(kv.LRange("list", 0, 4), bytes())
		require.ListLen(kv, "list", 0)

		require.Works(kv.LPush("funky2D", []byte{100, 255}))
		require.Works(kv.LPush("funky2D", []byte{0, 10}))
		require.AllEqual(kv.LRange("funky2D", 0, -1), [][]byte{{0, 10}, {100, 255}})
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		require := require{TB: t}

		// Strings group
		require.Works(kv.Set("empty-number", 0))
		require.Works(kv.Set("empty-string", ""))
		require.Works(kv.Set("empty-bytes", []byte{}))
		require.Equal(kv.Get("empty-number"), 0)
		require.Equal(kv.Get("empty-string"), "")
		require.Equal(kv.Get("empty-bytes"), "")

		// List group. Once empty we should be able to do a Get without a
		// wrongtype error.
		require.Works(kv.LPush("empty-list", "here today gone tomorrow"))
		require.Equal(kv.Get("empty-list"), errWrongType)
		require.Works(kv.LTrim("empty-list", -1, -2))
		require.Equal(kv.Get("empty-list"), nil)
	})

	t.Run("expire", func(t *testing.T) {
		// Skips because of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Parallel()

		require := require{TB: t}

		// Set removes expire
		{
			k := "expires-set-reset"
			require.Works(kv.SetEx(k, 60, "1"))
			require.Equal(kv.Get(k), "1")
			require.TTL(kv, k, 60)

			require.Works(kv.Set(k, "2"))
			require.Equal(kv.Get(k), "2")
			require.TTL(kv, k, -1)

		}

		// SetEx, Expire and TTL
		require.Works(kv.SetEx("expires-setex", 60, "1"))
		require.Works(kv.Set("expires-set", "1"))
		require.Works(kv.Expire("expires-set", 60))
		require.Works(kv.Set("expires-unset", "1"))
		require.TTL(kv, "expires-setex", 60)
		require.TTL(kv, "expires-set", 60)
		require.TTL(kv, "expires-unset", -1)
		require.TTL(kv, "expires-does-not-exist", -2)

		require.Equal(kv.Get("expires-setex"), "1")
		require.Equal(kv.Get("expires-set"), "1")

		require.Works(kv.SetEx("expires-setex", 1, "2"))
		require.Works(kv.Set("expires-set", "2"))
		require.Works(kv.Expire("expires-set", 1))

		time.Sleep(1100 * time.Millisecond)
		require.Equal(kv.Get("expires-setex"), nil)
		require.Equal(kv.Get("expires-set"), nil)
		require.TTL(kv, "expires-setex", -2)
		require.TTL(kv, "expires-set", -2)
	})

	t.Run("hash-expire", func(t *testing.T) {
		// Skips because of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Parallel()

		require := require{TB: t}

		// Hash mutations keep expire
		require.Works(kv.HSet("expires-unset-hash", "simple", "1"))
		require.Works(kv.HSet("expires-set-hash", "simple", "1"))
		require.Works(kv.Expire("expires-set-hash", 60))
		require.TTL(kv, "expires-unset-hash", -1)
		require.TTL(kv, "expires-set-hash", 60)
		require.Equal(kv.HGet("expires-unset-hash", "simple"), "1")
		require.Equal(kv.HGet("expires-set-hash", "simple"), "1")

		require.Works(kv.HSet("expires-unset-hash", "simple", "2"))
		require.Works(kv.HSet("expires-set-hash", "simple", "2"))
		require.TTL(kv, "expires-unset-hash", -1)
		require.TTL(kv, "expires-set-hash", 60)
		require.Equal(kv.HGet("expires-unset-hash", "simple"), "2")
		require.Equal(kv.HGet("expires-set-hash", "simple"), "2")

		// Check expiration happens on hashes
		require.Works(kv.Expire("expires-set-hash", 1))
		time.Sleep(1100 * time.Millisecond)
		require.Equal(kv.HGet("expires-set-hash", "simple"), nil)
		require.TTL(kv, "expires-set-hash", -2)
	})

	t.Run("hash-expire", func(t *testing.T) {
		// Skips because of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Parallel()

		require := require{TB: t}

		// Hash mutations keep expire
		require.Works(kv.LPush("expires-unset-list", "1"))
		require.Works(kv.LPush("expires-set-list", "1"))
		require.Works(kv.Expire("expires-set-list", 60))
		require.TTL(kv, "expires-unset-list", -1)
		require.TTL(kv, "expires-set-list", 60)
		require.AllEqual(kv.LRange("expires-unset-list", 0, -1), []string{"1"})
		require.AllEqual(kv.LRange("expires-set-list", 0, -1), []string{"1"})

		require.Works(kv.LPush("expires-unset-list", "2"))
		require.Works(kv.LPush("expires-set-list", "2"))
		require.TTL(kv, "expires-unset-list", -1)
		require.TTL(kv, "expires-set-list", 60)
		require.AllEqual(kv.LRange("expires-unset-list", 0, -1), []string{"2", "1"})
		require.AllEqual(kv.LRange("expires-set-list", 0, -1), []string{"2", "1"})

		// Check expiration happens on hashes
		require.Works(kv.Expire("expires-set-list", 1))
		time.Sleep(1100 * time.Millisecond)
		require.Equal(kv.HGet("expires-set-list", "simple"), nil)
		require.TTL(kv, "expires-set-list", -2)
	})

	t.Run("wrongtype", func(t *testing.T) {
		t.Parallel()

		require := require{TB: t}
		requireWrongType := func(err error) {
			t.Helper()
			if err == nil || !strings.Contains(err.Error(), "WRONGTYPE") {
				t.Fatalf("expected wrongtype error, got %v", err)
			}
		}

		require.Works(kv.Set("wrongtype-string", "1"))
		require.Works(kv.HSet("wrongtype-hash", "1", "1"))
		require.Works(kv.LPush("wrongtype-list", "1"))

		for _, k := range []string{"wrongtype-string", "wrongtype-hash", "wrongtype-list"} {
			// Ensure we fail Get when used against non string group
			if k != "wrongtype-string" {
				require.Equal(kv.Get(k), errWrongType)
				require.Equal(kv.GetSet(k, "2"), errWrongType)
				require.Equal(kv.Get(k), errWrongType) // ensure GetSet didn't set
				_, err := kv.Incr(k)
				requireWrongType(err)
			}

			// Ensure we fail hashes when used against non hashes.
			if k != "wrongtype-hash" {
				require.Equal(kv.HGet(k, "field"), errWrongType)
				require.Equal(redispool.Value(kv.HGetAll(k)), errWrongType)
				requireWrongType(kv.HSet(k, "field", "value"))
			}

			// Ensure we fail lists when used against non lists.
			if k != "wrongtype-list" {
				_, err := kv.LLen(k)
				requireWrongType(err)
				requireWrongType(kv.LPush(k, "1"))
				requireWrongType(kv.LTrim(k, 1, 2))
				require.Equal(redispool.Value(kv.LRange(k, 1, 2)), errWrongType)
			}

			// Ensure we can always override values with set
			require.Works(kv.Set(k, "2"))
			require.Equal(kv.Get(k), "2")
		}
	})
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

	if err := redispool.DeleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	kv := redispool.RedisKeyValue(pool).(interface {
		WithPrefix(string) redispool.KeyValue
	})
	return kv.WithPrefix(prefix)
}

func bytes(ss ...string) [][]byte {
	bs := make([][]byte, 0, len(ss))
	for _, s := range ss {
		bs = append(bs, []byte(s))
	}
	return bs
}

// require is redispool.Value helpers to make test readable
type require struct {
	testing.TB
}

func (t require) Works(err error) {
	// Works is a weird name, but it makes the function name align with Equal.
	t.Helper()
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
}

func (t require) Equal(got redispool.Value, want any) {
	t.Helper()
	switch wantV := want.(type) {
	case bool:
		gotV, err := got.Bool()
		t.Works(err)
		if gotV != wantV {
			t.Fatalf("got %v, wanted %v", gotV, wantV)
		}
	case []byte:
		gotV, err := got.Bytes()
		t.Works(err)
		if !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("got %q, wanted %q", gotV, wantV)
		}
	case int:
		gotV, err := got.Int()
		t.Works(err)
		if gotV != wantV {
			t.Fatalf("got %d, wanted %d", gotV, wantV)
		}
	case string:
		gotV, err := got.String()
		t.Works(err)
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
func (t require) AllEqual(got redispool.Values, want any) {
	t.Helper()
	switch wantV := want.(type) {
	case [][]byte:
		gotV, err := got.ByteSlices()
		t.Works(err)
		if !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("got %q, wanted %q", gotV, wantV)
		}
	case []string:
		gotV, err := got.Strings()
		t.Works(err)
		if !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("got %q, wanted %q", gotV, wantV)
		}
	case map[string]string:
		gotV, err := got.StringMap()
		t.Works(err)
		if !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("got %q, wanted %q", gotV, wantV)
		}
	default:
		t.Fatalf("unsupported want type for %q: %T", want, want)
	}
}
func (t require) ListLen(kv redispool.KeyValue, key string, want int) {
	t.Helper()
	got, err := kv.LLen(key)
	if err != nil {
		t.Fatal("LLen returned error", err)
	}
	if got != want {
		t.Fatalf("unexpected list length got=%d want=%d", got, want)
	}
}
func (t require) TTL(kv redispool.KeyValue, key string, want int) {
	t.Helper()
	got, err := kv.TTL(key)
	if err != nil {
		t.Fatal("TTL returned error", err)
	}

	// TTL timing is tough in a test environment. So if we are expecting a
	// positive TTL we give a 10s grace.
	if want > 10 {
		min := want - 10
		if got < min || got > want {
			t.Fatalf("unexpected TTL got=%d expected=[%d,%d]", got, min, want)
		}
	} else if want < 0 {
		if got != want {
			t.Fatalf("unexpected TTL got=%d want=%d", got, want)
		}
	} else {
		t.Fatalf("got bad want value %d", want)
	}
}
