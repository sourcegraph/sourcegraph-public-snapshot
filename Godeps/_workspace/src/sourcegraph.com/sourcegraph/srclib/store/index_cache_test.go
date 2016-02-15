package store

import (
	"container/list"
	"fmt"
	"testing"
)

type mockCacheableIndexStore struct{}

func (m *mockCacheableIndexStore) StoreKey() interface{} { return "test" }

type mockIndex struct {
	id int
}

func (m *mockIndex) Ready() bool              { return true }
func (m *mockIndex) Covers(f interface{}) int { return 1 }

func TestCache(t *testing.T) {
	store := &mockCacheableIndexStore{}
	index1 := &mockIndex{123}
	index2 := &mockIndex{456}

	// empty cache, should use fallback
	if index1 != cacheGet(store, "test_index", index1) {
		t.Errorf("cacheGet expected to use fallback value")
	}

	// We put in 2, and get with fallback 1. We should get back 2
	cachePut(store, "test_index", index2)
	if index2 != cacheGet(store, "test_index", index1) {
		t.Errorf("cachePut followed by cacheGet returns different results")
	}

	// Same test, but on different key with indexes swapped
	if index2 != cacheGet(store, "test_index_2", index2) {
		t.Errorf("cacheGet expected to use fallback value")
	}
	cachePut(store, "test_index_2", index1)
	if index1 != cacheGet(store, "test_index_2", index2) {
		t.Errorf("cachePut followed by cacheGet returns different results")
	}

}

func TestLRU(t *testing.T) {
	store := &mockCacheableIndexStore{}
	cacheSize := 50
	c := &indexCache{
		indexes: map[indexCacheKey]*list.Element{},
		lru:     list.New(),
		maxLen:  cacheSize,
	}
	for i := 0; i < cacheSize+5; i++ {
		index := &mockIndex{i}
		c.cachePut(store, fmt.Sprintf("index_%d", i), index)
	}

	// Now indexes < 5 should have been evicted, everything else should still be there
	fallback := &mockIndex{cacheSize * 2}
	for i := 0; i < cacheSize+5; i++ {
		index := c.cacheGet(store, fmt.Sprintf("index_%d", i), fallback)
		if i < 5 && index != fallback {
			t.Errorf("index_%d should have been evicted", i)
		} else if i >= 5 && index == fallback {
			t.Errorf("index_%d should not have been evicted", i)
		}
	}

	// Do some NOOP puts. In our implementation a NOOP put does not affect
	// LRU
	for i := cacheSize / 3; i < cacheSize/2; i++ {
		c.cachePut(store, fmt.Sprintf("index_%d", i), fallback)
	}
	for i := 5; i < cacheSize+5; i++ {
		if fallback == c.cacheGet(store, fmt.Sprintf("index_%d", i), fallback) {
			t.Errorf("A NOOP put on index_%d updated the cache", i)
		}
	}

	// The LRU index we did a get on was index_5. Do a put and ensure it
	// is gone
	c.cachePut(store, fmt.Sprintf("index_%d", cacheSize+5), &mockIndex{cacheSize + 5})
	for i := 0; i <= cacheSize+5; i++ {
		index := c.cacheGet(store, fmt.Sprintf("index_%d", i), fallback)
		if i <= 5 && index != fallback {
			t.Errorf("index_%d should have been evicted", i)
		} else if i > 5 && index == fallback {
			t.Errorf("index_%d should not have been evicted", i)
		}
	}
}
