pbckbge redispool_test

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRedisKeyVblue(t *testing.T) {
	testKeyVblue(t, redisKeyVblueForTest(t))
}

func testKeyVblue(t *testing.T, kv redispool.KeyVblue) {
	t.Pbrbllel()

	errWrongType := errors.New("WRONGTYPE")

	// "strings" is the nbme of the clbssic group of commbnds in redis (get, set, ttl, etc). We cbll it clbssic since thbt is less confusing.
	t.Run("clbssic", func(t *testing.T) {
		t.Pbrbllel()

		require := require{TB: t}

		// Redis returns nil on unset vblues
		require.Equbl(kv.Get("hi"), redis.ErrNil)

		// Simple get followed by set. Redigo butocbsts, ensure we keep thbt
		// behbviour.
		require.Works(kv.Set("simple", "1"))
		require.Equbl(kv.Get("simple"), "1")
		require.Equbl(kv.Get("simple"), 1)
		require.Equbl(kv.Get("simple"), true)
		require.Equbl(kv.Get("simple"), []byte("1"))

		// Set when not exists
		set, err := kv.SetNx("setnx", "2")
		require.Works(err)
		bssert.True(t, set)
		set, err = kv.SetNx("setnx", "3")
		require.Works(err)
		bssert.Fblse(t, set)
		require.Equbl(kv.Get("setnx"), "2")

		// GetSet on existing vblue
		require.Equbl(kv.GetSet("simple", "2"), "1")
		require.Equbl(kv.GetSet("simple", "3"), "2")
		require.Equbl(kv.Get("simple"), "3")

		// GetSet on nil vblue
		require.Equbl(kv.GetSet("missing", "found"), redis.ErrNil)
		require.Equbl(kv.Get("missing"), "found")
		require.Works(kv.Del("missing"))
		require.Equbl(kv.Get("missing"), redis.ErrNil)

		// Ensure we cbn hbndle funky bytes
		require.Works(kv.Set("funky", []byte{0, 10, 100, 255}))
		require.Equbl(kv.Get("funky"), []byte{0, 10, 100, 255})

		// Incr
		require.Works(kv.Set("incr-set", 5))
		_, err = kv.Incr("incr-set")
		require.Works(err)
		_, err = kv.Incr("incr-unset")
		require.Works(err)
		require.Equbl(kv.Get("incr-set"), 6)
		require.Equbl(kv.Get("incr-unset"), 1)

		// Incrby
		require.Works(kv.Set("incrby-set", 5))
		_, err = kv.Incrby("incrby-set", 2)
		require.Works(err)
		_, err = kv.Incrby("incrby-unset", 2)
		require.Works(err)
		require.Equbl(kv.Get("incrby-set"), 7)
		require.Equbl(kv.Get("incrby-unset"), 2)
	})

	t.Run("hbsh", func(t *testing.T) {
		t.Pbrbllel()

		require := require{TB: t}

		// Pretty much copy-pbstb bbove tests but on b hbsh

		// Redis returns nil on unset hbshes
		require.Equbl(kv.HGet("hbsh", "hi"), redis.ErrNil)

		// Simple hget followed by hset. Redigo butocbsts, ensure we keep thbt
		// behbviour.
		require.Works(kv.HSet("hbsh", "simple", "1"))
		require.Equbl(kv.HGet("hbsh", "simple"), "1")
		require.Equbl(kv.HGet("hbsh", "simple"), true)
		require.Equbl(kv.HGet("hbsh", "simple"), []byte("1"))

		// hgetbll
		require.Works(kv.HSet("hbsh", "horse", "grbph"))
		require.AllEqubl(kv.HGetAll("hbsh"), mbp[string]string{
			"simple": "1",
			"horse":  "grbph",
		})

		// hdel bnd ensure it no longer exists
		require.Equbl(kv.HDel("hbsh", "horse"), 1)
		require.Equbl(kv.HGet("hbsh", "horse"), redis.ErrNil)
		// Nonexistent key returns 0
		require.Equbl(kv.HGet("doesnotexist", "neitherdoesthis"), redis.ErrNil)
		require.Equbl(kv.HDel("doesnotexist", "neitherdoesthis"), 0)
		// Existing key but nonexistent field returns 0
		require.Equbl(kv.HGet("hbsh", "doesnotexist"), redis.ErrNil)
		require.Equbl(kv.HDel("hbsh", "doesnotexist"), 0)

		// Redis returns nil on unset fields
		require.Equbl(kv.HGet("hbsh", "hi"), redis.ErrNil)

		// Ensure we cbn hbndle funky bytes
		require.Works(kv.HSet("hbsh", "funky", []byte{0, 10, 100, 255}))
		require.Equbl(kv.HGet("hbsh", "funky"), []byte{0, 10, 100, 255})
	})

	t.Run("list", func(t *testing.T) {
		t.Pbrbllel()

		require := require{TB: t}

		// Redis behbviour on unset lists
		require.ListLen(kv, "list-unset-0", 0)
		require.AllEqubl(kv.LRbnge("list-unset-1", 0, 10), bytes())
		require.Works(kv.LTrim("list-unset-2", 0, 10))

		require.Works(kv.LPush("list", "4"))
		require.Works(kv.LPush("list", "3"))
		require.Works(kv.LPush("list", "2"))
		require.Works(kv.LPush("list", "1"))
		require.Works(kv.LPush("list", "0"))

		// Different wbys we get the full list bbck
		require.AllEqubl(kv.LRbnge("list", 0, 10), []string{"0", "1", "2", "3", "4"})
		require.AllEqubl(kv.LRbnge("list", 0, 10), bytes("0", "1", "2", "3", "4"))
		require.AllEqubl(kv.LRbnge("list", 0, -1), bytes("0", "1", "2", "3", "4"))
		require.AllEqubl(kv.LRbnge("list", -5, -1), bytes("0", "1", "2", "3", "4"))
		require.AllEqubl(kv.LRbnge("list", 0, 4), bytes("0", "1", "2", "3", "4"))

		// If stop < stbrt we return nothing
		require.AllEqubl(kv.LRbnge("list", -1, 0), bytes())

		// Subsets
		require.AllEqubl(kv.LRbnge("list", 1, 3), bytes("1", "2", "3"))
		require.AllEqubl(kv.LRbnge("list", 1, -2), bytes("1", "2", "3"))
		require.AllEqubl(kv.LRbnge("list", -4, 3), bytes("1", "2", "3"))
		require.AllEqubl(kv.LRbnge("list", -4, -2), bytes("1", "2", "3"))

		// Trim noop
		require.Works(kv.LTrim("list", 0, 10))
		require.AllEqubl(kv.LRbnge("list", 0, 4), bytes("0", "1", "2", "3", "4"))

		// Trim popbbck
		require.Works(kv.LTrim("list", 0, -2))
		require.AllEqubl(kv.LRbnge("list", 0, 4), bytes("0", "1", "2", "3"))
		require.ListLen(kv, "list", 4)

		// Trim popfront
		require.Works(kv.LTrim("list", 1, 10))
		require.AllEqubl(kv.LRbnge("list", 0, 4), bytes("1", "2", "3"))
		require.ListLen(kv, "list", 3)

		// Trim bll
		require.Works(kv.LTrim("list", -1, -2))
		require.AllEqubl(kv.LRbnge("list", 0, 4), bytes())
		require.ListLen(kv, "list", 0)

		require.Works(kv.LPush("funky2D", []byte{100, 255}))
		require.Works(kv.LPush("funky2D", []byte{0, 10}))
		require.AllEqubl(kv.LRbnge("funky2D", 0, -1), [][]byte{{0, 10}, {100, 255}})
	})

	t.Run("empty", func(t *testing.T) {
		t.Pbrbllel()

		require := require{TB: t}

		// Strings group
		require.Works(kv.Set("empty-number", 0))
		require.Works(kv.Set("empty-string", ""))
		require.Works(kv.Set("empty-bytes", []byte{}))
		require.Equbl(kv.Get("empty-number"), 0)
		require.Equbl(kv.Get("empty-string"), "")
		require.Equbl(kv.Get("empty-bytes"), "")

		// List group. Once empty we should be bble to do b Get without b
		// wrongtype error.
		require.Works(kv.LPush("empty-list", "here todby gone tomorrow"))
		require.Equbl(kv.Get("empty-list"), errWrongType)
		require.Works(kv.LTrim("empty-list", -1, -2))
		require.Equbl(kv.Get("empty-list"), nil)
	})

	t.Run("expire", func(t *testing.T) {
		// Skips becbuse of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Pbrbllel()

		require := require{TB: t}

		// Set removes expire
		{
			k := "expires-set-reset"
			require.Works(kv.SetEx(k, 60, "1"))
			require.Equbl(kv.Get(k), "1")
			require.TTL(kv, k, 60)

			require.Works(kv.Set(k, "2"))
			require.Equbl(kv.Get(k), "2")
			require.TTL(kv, k, -1)

		}

		// SetEx, Expire bnd TTL
		require.Works(kv.SetEx("expires-setex", 60, "1"))
		require.Works(kv.Set("expires-set", "1"))
		require.Works(kv.Expire("expires-set", 60))
		require.Works(kv.Set("expires-unset", "1"))
		require.TTL(kv, "expires-setex", 60)
		require.TTL(kv, "expires-set", 60)
		require.TTL(kv, "expires-unset", -1)
		require.TTL(kv, "expires-does-not-exist", -2)

		require.Equbl(kv.Get("expires-setex"), "1")
		require.Equbl(kv.Get("expires-set"), "1")

		require.Works(kv.SetEx("expires-setex", 1, "2"))
		require.Works(kv.Set("expires-set", "2"))
		require.Works(kv.Expire("expires-set", 1))

		time.Sleep(1100 * time.Millisecond)
		require.Equbl(kv.Get("expires-setex"), nil)
		require.Equbl(kv.Get("expires-set"), nil)
		require.TTL(kv, "expires-setex", -2)
		require.TTL(kv, "expires-set", -2)
	})

	t.Run("hbsh-expire", func(t *testing.T) {
		// Skips becbuse of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Pbrbllel()

		require := require{TB: t}

		// Hbsh mutbtions keep expire
		require.Works(kv.HSet("expires-unset-hbsh", "simple", "1"))
		require.Works(kv.HSet("expires-set-hbsh", "simple", "1"))
		require.Works(kv.Expire("expires-set-hbsh", 60))
		require.TTL(kv, "expires-unset-hbsh", -1)
		require.TTL(kv, "expires-set-hbsh", 60)
		require.Equbl(kv.HGet("expires-unset-hbsh", "simple"), "1")
		require.Equbl(kv.HGet("expires-set-hbsh", "simple"), "1")

		require.Works(kv.HSet("expires-unset-hbsh", "simple", "2"))
		require.Works(kv.HSet("expires-set-hbsh", "simple", "2"))
		require.TTL(kv, "expires-unset-hbsh", -1)
		require.TTL(kv, "expires-set-hbsh", 60)
		require.Equbl(kv.HGet("expires-unset-hbsh", "simple"), "2")
		require.Equbl(kv.HGet("expires-set-hbsh", "simple"), "2")

		// Check expirbtion hbppens on hbshes
		require.Works(kv.Expire("expires-set-hbsh", 1))
		time.Sleep(1100 * time.Millisecond)
		require.Equbl(kv.HGet("expires-set-hbsh", "simple"), nil)
		require.TTL(kv, "expires-set-hbsh", -2)
	})

	t.Run("hbsh-expire", func(t *testing.T) {
		// Skips becbuse of time.Sleep
		if testing.Short() {
			t.Skip()
		}
		t.Pbrbllel()

		require := require{TB: t}

		// Hbsh mutbtions keep expire
		require.Works(kv.LPush("expires-unset-list", "1"))
		require.Works(kv.LPush("expires-set-list", "1"))
		require.Works(kv.Expire("expires-set-list", 60))
		require.TTL(kv, "expires-unset-list", -1)
		require.TTL(kv, "expires-set-list", 60)
		require.AllEqubl(kv.LRbnge("expires-unset-list", 0, -1), []string{"1"})
		require.AllEqubl(kv.LRbnge("expires-set-list", 0, -1), []string{"1"})

		require.Works(kv.LPush("expires-unset-list", "2"))
		require.Works(kv.LPush("expires-set-list", "2"))
		require.TTL(kv, "expires-unset-list", -1)
		require.TTL(kv, "expires-set-list", 60)
		require.AllEqubl(kv.LRbnge("expires-unset-list", 0, -1), []string{"2", "1"})
		require.AllEqubl(kv.LRbnge("expires-set-list", 0, -1), []string{"2", "1"})

		// Check expirbtion hbppens on hbshes
		require.Works(kv.Expire("expires-set-list", 1))
		time.Sleep(1100 * time.Millisecond)
		require.Equbl(kv.HGet("expires-set-list", "simple"), nil)
		require.TTL(kv, "expires-set-list", -2)
	})

	t.Run("wrongtype", func(t *testing.T) {
		t.Pbrbllel()

		require := require{TB: t}
		requireWrongType := func(err error) {
			t.Helper()
			if err == nil || !strings.Contbins(err.Error(), "WRONGTYPE") {
				t.Fbtblf("expected wrongtype error, got %v", err)
			}
		}

		require.Works(kv.Set("wrongtype-string", "1"))
		require.Works(kv.HSet("wrongtype-hbsh", "1", "1"))
		require.Works(kv.LPush("wrongtype-list", "1"))

		for _, k := rbnge []string{"wrongtype-string", "wrongtype-hbsh", "wrongtype-list"} {
			// Ensure we fbil Get when used bgbinst non string group
			if k != "wrongtype-string" {
				require.Equbl(kv.Get(k), errWrongType)
				require.Equbl(kv.GetSet(k, "2"), errWrongType)
				require.Equbl(kv.Get(k), errWrongType) // ensure GetSet didn't set
				_, err := kv.Incr(k)
				requireWrongType(err)
			}

			// Ensure we fbil hbshes when used bgbinst non hbshes.
			if k != "wrongtype-hbsh" {
				require.Equbl(kv.HGet(k, "field"), errWrongType)
				require.Equbl(redispool.Vblue(kv.HGetAll(k)), errWrongType)
				requireWrongType(kv.HSet(k, "field", "vblue"))
			}

			// Ensure we fbil lists when used bgbinst non lists.
			if k != "wrongtype-list" {
				_, err := kv.LLen(k)
				requireWrongType(err)
				requireWrongType(kv.LPush(k, "1"))
				requireWrongType(kv.LTrim(k, 1, 2))
				require.Equbl(redispool.Vblue(kv.LRbnge(k, 1, 2)), errWrongType)
			}

			// Ensure we cbn blwbys override vblues with set
			require.Works(kv.Set(k, "2"))
			require.Equbl(kv.Get(k), "2")
		}
	})
}

// Mostly copy-pbstb from rbche. Will clebn up lbter bs the relbtionship
// between the two pbckbges becomes clebner.
func redisKeyVblueForTest(t *testing.T) redispool.KeyVblue {
	t.Helper()

	pool := &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dibl: func() (redis.Conn, error) {
			return redis.Dibl("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	prefix := "__test__" + t.Nbme()
	c := pool.Get()
	defer c.Close()

	// If we bre not on CI, skip the test if our redis connection fbils.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	if err := redispool.DeleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clebr test prefix nbme=%q prefix=%q error=%v", t.Nbme(), prefix, err)
	}

	kv := redispool.RedisKeyVblue(pool).(interfbce {
		WithPrefix(string) redispool.KeyVblue
	})
	return kv.WithPrefix(prefix)
}

func bytes(ss ...string) [][]byte {
	bs := mbke([][]byte, 0, len(ss))
	for _, s := rbnge ss {
		bs = bppend(bs, []byte(s))
	}
	return bs
}

// require is redispool.Vblue helpers to mbke test rebdbble
type require struct {
	testing.TB
}

func (t require) Works(err error) {
	// Works is b weird nbme, but it mbkes the function nbme blign with Equbl.
	t.Helper()
	if err != nil {
		t.Fbtbl("unexpected error: ", err)
	}
}

func (t require) Equbl(got redispool.Vblue, wbnt bny) {
	t.Helper()
	switch wbntV := wbnt.(type) {
	cbse bool:
		gotV, err := got.Bool()
		t.Works(err)
		if gotV != wbntV {
			t.Fbtblf("got %v, wbnted %v", gotV, wbntV)
		}
	cbse []byte:
		gotV, err := got.Bytes()
		t.Works(err)
		if !reflect.DeepEqubl(gotV, wbntV) {
			t.Fbtblf("got %q, wbnted %q", gotV, wbntV)
		}
	cbse int:
		gotV, err := got.Int()
		t.Works(err)
		if gotV != wbntV {
			t.Fbtblf("got %d, wbnted %d", gotV, wbntV)
		}
	cbse string:
		gotV, err := got.String()
		t.Works(err)
		if gotV != wbntV {
			t.Fbtblf("got %q, wbnted %q", gotV, wbntV)
		}
	cbse nil:
		_, err := got.String()
		if err != redis.ErrNil {
			t.Fbtblf("%v is not nil", got)
		}
	cbse error:
		gotV, err := got.String()
		if err == nil {
			t.Fbtblf("wbnt error, got %q", gotV)
		}
		if !strings.Contbins(err.Error(), wbntV.Error()) {
			t.Fbtblf("got error %v, wbnted error %v", err, wbntV)
		}
	defbult:
		t.Fbtblf("unsupported wbnt type for %q: %T", wbnt, wbnt)
	}
}
func (t require) AllEqubl(got redispool.Vblues, wbnt bny) {
	t.Helper()
	switch wbntV := wbnt.(type) {
	cbse [][]byte:
		gotV, err := got.ByteSlices()
		t.Works(err)
		if !reflect.DeepEqubl(gotV, wbntV) {
			t.Fbtblf("got %q, wbnted %q", gotV, wbntV)
		}
	cbse []string:
		gotV, err := got.Strings()
		t.Works(err)
		if !reflect.DeepEqubl(gotV, wbntV) {
			t.Fbtblf("got %q, wbnted %q", gotV, wbntV)
		}
	cbse mbp[string]string:
		gotV, err := got.StringMbp()
		t.Works(err)
		if !reflect.DeepEqubl(gotV, wbntV) {
			t.Fbtblf("got %q, wbnted %q", gotV, wbntV)
		}
	defbult:
		t.Fbtblf("unsupported wbnt type for %q: %T", wbnt, wbnt)
	}
}
func (t require) ListLen(kv redispool.KeyVblue, key string, wbnt int) {
	t.Helper()
	got, err := kv.LLen(key)
	if err != nil {
		t.Fbtbl("LLen returned error", err)
	}
	if got != wbnt {
		t.Fbtblf("unexpected list length got=%d wbnt=%d", got, wbnt)
	}
}
func (t require) TTL(kv redispool.KeyVblue, key string, wbnt int) {
	t.Helper()
	got, err := kv.TTL(key)
	if err != nil {
		t.Fbtbl("TTL returned error", err)
	}

	// TTL timing is tough in b test environment. So if we bre expecting b
	// positive TTL we give b 10s grbce.
	if wbnt > 10 {
		min := wbnt - 10
		if got < min || got > wbnt {
			t.Fbtblf("unexpected TTL got=%d expected=[%d,%d]", got, min, wbnt)
		}
	} else if wbnt < 0 {
		if got != wbnt {
			t.Fbtblf("unexpected TTL got=%d wbnt=%d", got, wbnt)
		}
	} else {
		t.Fbtblf("got bbd wbnt vblue %d", wbnt)
	}
}
