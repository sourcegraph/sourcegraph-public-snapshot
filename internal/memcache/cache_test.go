package memcache

import (
	"reflect"
	"sync/atomic"
	"testing"
)

func TestLRU(t *testing.T) {
	cache, err := New(64)
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	numCacheMisses := 0
	assertKeyValue := func(key interface{}) {
		value, err := cache.GetOrCreate(key, func() (interface{}, int, error) {
			numCacheMisses++
			return key, 1, nil
		})
		if err != nil {
			t.Fatalf("unexpected error from GetOrCreate: %s", err)
		}
		if value != key {
			t.Errorf("unexpected cached value: want=%d have=%d", key, value)
		}
	}

	// Fresh values
	for i := 0; i < 128; i++ {
		assertKeyValue(i)
	}
	if numCacheMisses != 128 {
		t.Errorf("unexpected number of cache misses: want=%d have=%d", 128, numCacheMisses)
	}

	// Cached values
	for i := 64; i < 128; i++ {
		assertKeyValue(i)
	}
	if numCacheMisses != 128 {
		t.Errorf("unexpected number of cache misses: want=%d have=%d", 128, numCacheMisses)
	}

	// Evicted values
	for i := 0; i < 64; i++ {
		assertKeyValue(i)
	}
	if numCacheMisses != 192 {
		t.Errorf("unexpected number of cache misses: want=%d have=%d", 192, numCacheMisses)
	}
}

func TestEntrySize(t *testing.T) {
	var evicted []interface{}
	cache, err := NewWithEvict(50, func(key interface{}, value interface{}) {
		evicted = append(evicted, key)
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	numCacheMisses := 0
	assertKeyValue := func(key interface{}, size int) {
		value, err := cache.GetOrCreate(key, func() (interface{}, int, error) {
			numCacheMisses++
			return key, size, nil
		})
		if err != nil {
			t.Fatalf("unexpected error from GetOrCreate: %s", err)
		}
		if value != key {
			t.Errorf("unexpected cached value: want=%d have=%d", key, value)
		}
	}

	assertKeyValue("foo", 25)  // size = 25
	assertKeyValue("bar", 25)  // size = 50
	assertKeyValue("baz", 27)  // size = 77; evict "foo" and "bar"; size = 27
	assertKeyValue("foo", 10)  // size = 37
	assertKeyValue("bar", 10)  // size = 47
	assertKeyValue("s01", 1)   // size = 48
	assertKeyValue("s02", 1)   // size = 49
	assertKeyValue("s03", 1)   // size = 50
	assertKeyValue("s04", 1)   // size = 51; evict "baz"; size = 24
	assertKeyValue("bonk", 26) // size = 50

	if numCacheMisses != 10 {
		t.Errorf("unexpected number of cache misses: want=%d have=%d", 10, numCacheMisses)
	}

	expected := []interface{}{"foo", "bar", "baz"}
	if !reflect.DeepEqual(evicted, expected) {
		t.Errorf("unexpected evictions: want=%s have=%s", expected, evicted)
	}
}

func TestGetOrCreateConcurrentFactoryInvocations(t *testing.T) {
	// This tests a race that can happen between acquiring the locks of
	// cache.get and cache.add. We do not hold a singel lock as the factory
	// function may be expensive and the lock would be too coarse, blocking
	// all cache reads while the factory is in its critical section.
	//
	// This test ensures that two concurrent invocations of cache.GetOrCreate
	// acquire the locks in this order:
	//
	//   1. [call A].get
	//   2. [call B].get
	//   3. [call A].add
	//   4. [call B].add // evicts call A's value
	//   5. [call C].get // gets call B's value

	var evicted []interface{}
	cache, err := NewWithEvict(50, func(key interface{}, value interface{}) {
		evicted = append(evicted, value)
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	numCacheMisses := uint64(0)
	assertKeyValue := func(key, expectedValue interface{}, fu func() interface{}) {
		value, err := cache.GetOrCreate(key, func() (interface{}, int, error) {
			atomic.AddUint64(&numCacheMisses, 1)
			return fu(), 1, nil
		})
		if err != nil {
			t.Fatalf("unexpected error from GetOrCreate: %s", err)
		}
		if value != expectedValue {
			t.Errorf("unexpected cached value: want=%s have=%s", expectedValue, value)
		}
	}

	tick := make(chan struct{})
	defer close(tick)

	assertKeyValueViaChannel := func(key, expectedValue interface{}, factoryValueCh <-chan interface{}) {
		assertKeyValue(key, expectedValue, func() interface{} {
			tick <- struct{}{}      // sync point 1
			return <-factoryValueCh // blocked by main test
		})

		tick <- struct{}{} // sync point 2
	}

	// concurrent call A
	val1 := make(chan interface{})
	defer close(val1)
	go assertKeyValueViaChannel("foo", "bar", val1) // gets own value

	// concurrent call B
	val2 := make(chan interface{})
	defer close(val2)
	go assertKeyValueViaChannel("foo", "baz", val2) // gets own value

	<-tick        // wait for A (sync point 1)
	<-tick        // wait for B (sync point 1)
	val1 <- "bar" // release A
	<-tick        // wait for A (sync point 2)
	val2 <- "baz" // release B
	<-tick        // wait for B (sync point 2)

	// call C
	assertKeyValue("foo", "baz", func() interface{} {
		return "bonk"
	})

	if numCacheMisses != 2 {
		t.Errorf("unexpected number of cache misses: want=%d have=%d", 2, numCacheMisses)
	}

	expected := []interface{}{"bar"}
	if !reflect.DeepEqual(evicted, expected) {
		t.Errorf("unexpected evictions: want=%s have=%s", expected, evicted)
	}
}
