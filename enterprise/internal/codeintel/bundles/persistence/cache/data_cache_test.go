package cache

import (
	"testing"
)

type dataCacheTestCase struct {
	key   string
	value interface{}
}

func TestDataCache(t *testing.T) {
	cache, err := NewDataCache(10)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %s", err)
	}

	_ = cache.Set("foo", 123, 5)
	_ = cache.Set("bar", 234, 4)
	_ = cache.Set("baz", 345, 1)

	assertDataCache(t, cache, []dataCacheTestCase{
		{"foo", 123},
		{"bar", 234},
		{"baz", 345},
	})

	// Cache at max capacity
	_ = cache.Set("bonk", 456, 2)

	assertDataCache(t, cache, []dataCacheTestCase{
		{"bonk", 456},
	})

	count := 0
	for _, key := range []string{"foo", "bar", "baz"} {
		if _, ok := cache.Get(key); ok {
			count++
		}
	}

	if count == 3 {
		t.Errorf("expected an eviction")
	}
}

func assertDataCache(t *testing.T, cache DataCache, testCases []dataCacheTestCase) {
	waitForCondition(func() bool {
		for _, testCase := range testCases {
			if _, ok := cache.Get(testCase.key); !ok {
				return false
			}
		}

		return true
	})

	for _, testCase := range testCases {
		value, ok := cache.Get(testCase.key)
		if !ok {
			t.Errorf("expected %v to exist in cache", testCase.key)
		} else if value != testCase.value {
			t.Errorf("unexpected value. want=%v have=%v", testCase.value, value)
		}
	}
}
