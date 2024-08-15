package rcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/httpcache"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

func TestCache_namespace(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	type testcase struct {
		prefix  string
		entries map[string]string
	}

	cases := []testcase{
		{
			prefix: "a",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "b",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "c",
			entries: map[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		},
	}

	caches := make([]httpcache.Cache, len(cases))
	for i, test := range cases {
		caches[i] = New(kv, test.prefix)
		for k, v := range test.entries {
			caches[i].Set(ctx, k, []byte(v))
		}
	}
	for i, test := range cases {
		// test all the keys that should be present are found
		for k, v := range test.entries {
			b, ok := caches[i].Get(ctx, k)
			if !ok {
				t.Fatalf("error getting entry from redis (prefix=%s)", test.prefix)
			}
			if string(b) != v {
				t.Errorf("expected %s, got %s", v, string(b))
			}
		}

		// test not found case
		if _, ok := caches[i].Get(ctx, "not-found"); ok {
			t.Errorf("expected not found")
		}
	}
}

func TestCache_simple(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	c := New(kv, "some_prefix")
	_, ok := c.Get(ctx, "a")
	if ok {
		t.Fatal("Initial Get should find nothing")
	}

	c.Set(ctx, "a", []byte("b"))
	b, ok := c.Get(ctx, "a")
	if !ok {
		t.Fatal("Expect to get a after setting")
	}
	if string(b) != "b" {
		t.Fatalf("got %v, want %v", string(b), "b")
	}

	c.Delete(ctx, "a")
	_, ok = c.Get(ctx, "a")
	if ok {
		t.Fatal("Get after delete should of found nothing")
	}
}

func TestCache_Increase(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	c := NewWithTTL(kv, "some_prefix", 1)
	c.Increase(ctx, "a")

	got, ok := c.Get(ctx, "a")
	assert.True(t, ok)
	assert.Equal(t, []byte("1"), got)

	time.Sleep(time.Second)

	// now wait upto another 5s. We do this because timing is hard.
	assert.Eventually(t, func() bool {
		_, ok = c.Get(ctx, "a")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcache.increase did not respect expiration")
}

func TestCache_KeyTTL(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	c := NewWithTTL(kv, "some_prefix", 1)
	c.Set(ctx, "a", []byte("b"))

	ttl, ok := c.KeyTTL(ctx, "a")
	assert.True(t, ok)
	assert.Equal(t, 1, ttl)

	time.Sleep(time.Second)

	// now wait upto another 5s. We do this because timing is hard.
	assert.Eventually(t, func() bool {
		_, ok = c.KeyTTL(ctx, "a")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcache.ketttl did not respect expiration")

	c.SetWithTTL(ctx, "c", []byte("d"), 0) // invalid TTL
	_, ok = c.KeyTTL(ctx, "c")
	if ok {
		t.Fatal("KeyTTL after setting invalid ttl should have found nothing")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	c := NewWithTTL(kv, "some_prefix", 60)
	c.SetWithTTL(ctx, "a", []byte("b"), 30)
	b, ok := c.Get(ctx, "a")
	if !ok {
		t.Fatal("Expect to get a after setting")
	}
	if string(b) != "b" {
		t.Fatalf("got %v, want %v", string(b), "b")
	}
	ttl, ok := c.KeyTTL(ctx, "a")
	if !ok {
		t.Fatal("Expect to be able to read ttl after setting")
	}
	if ttl > 30 {
		t.Fatalf("ttl got %v, want %v", ttl, 30)
	}

	c.Delete(ctx, "a")
	_, ok = c.Get(ctx, "a")
	if ok {
		t.Fatal("Get after delete should have found nothing")
	}

	c.SetWithTTL(ctx, "c", []byte("d"), 0) // invalid operation
	_, ok = c.Get(ctx, "c")
	if ok {
		t.Fatal("SetWithTTL should not create a key with invalid expiry")
	}
}

func TestCache_Hashes(t *testing.T) {
	ctx := tenant.TestContext()
	kv := SetupForTest(t)

	// Test SetHashItem
	c := NewWithTTL(kv, "simple_hash", 1)
	err := c.SetHashItem(ctx, "key", "hashKey1", "value1")
	assert.NoError(t, err)
	err = c.SetHashItem(ctx, "key", "hashKey2", "value2")
	assert.NoError(t, err)

	// Test GetHashItem
	val1, err := c.GetHashItem(ctx, "key", "hashKey1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val1)
	val2, err := c.GetHashItem(ctx, "key", "hashKey2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val2)
	val3, err := c.GetHashItem(ctx, "key", "hashKey3")
	assert.Error(t, err)
	assert.Equal(t, "", val3)

	// Test GetHashAll
	all, err := c.GetHashAll(ctx, "key")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"hashKey1": "value1", "hashKey2": "value2"}, all)

	// Test DeleteHashItem
	// Bit redundant, but double check that the key still exists
	val1, err = c.GetHashItem(ctx, "key", "hashKey1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val1)
	del1, err := c.DeleteHashItem(ctx, "key", "hashKey1")
	assert.NoError(t, err)
	assert.Equal(t, 1, del1)
	// Verify that it no longer exists
	val1, err = c.GetHashItem(ctx, "key", "hashKey1")
	assert.Error(t, err)
	assert.Equal(t, "", val1)
	// Delete nonexistent field: should return 0 (represents deleted items)
	val3, err = c.GetHashItem(ctx, "key", "hashKey3")
	assert.Error(t, err)
	assert.Equal(t, "", val3)
	del3, err := c.DeleteHashItem(ctx, "key", "hashKey3")
	assert.NoError(t, err)
	assert.Equal(t, 0, del3)
	// Delete nonexistent key: should return 0 (represents deleted items)
	val4, err := c.GetHashItem(ctx, "nonexistentkey", "nonexistenthashkey")
	assert.Error(t, err)
	assert.Equal(t, "", val4)
	del4, err := c.DeleteHashItem(ctx, "nonexistentkey", "nonexistenthashkey")
	assert.NoError(t, err)
	assert.Equal(t, 0, del4)
}
