package ttlcache

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// withClock sets the clock to be used by the cache. This is useful for testing.
func withClock[K comparable, V any](clock clock) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.clock = clock
	}
}

func TestGet(t *testing.T) {
	callCount := 0
	newEntryFunc := func(k string) int {
		callCount++
		return len(k)
	}

	options := []Option[string, int]{
		WithTTL[string, int](24 * time.Hour), // more than enough time for no expirations to occur
	}

	cache := New(newEntryFunc, options...)

	// Test that the cache returns the correct value for a key that has been added.
	value := cache.Get("hello")
	if value != 5 {
		t.Errorf("expected cache to return 5, got %d", value)
	}

	// Test that newEntryFunc was called once for the new key.
	if callCount != 1 {
		t.Errorf("expected newEntryFunc to be called once, got %d", callCount)
	}

	// Test that the cache returns the same value for the same key.
	value2 := cache.Get("hello")
	if value2 != 5 {
		t.Errorf("expected cache to return 5, got %d", value2)
	}

	// Test that the cache does not call newEntryFunc for an existing key.
	if callCount != 1 {
		t.Errorf("expected newEntryFunc to be called only once, got %d", callCount)
	}

	// Test that the cache returns a different value for a different key.
	value3 := cache.Get("foo")
	if value3 != 3 {
		t.Errorf("expected cache to return 3, got %d", value3)
	}

	// Test that newEntryFunc was called again for the new key.
	if callCount != 2 {
		t.Errorf("expected newEntryFunc to be called twice, got %d", callCount)
	}
}

func TestExpiration_Series(t *testing.T) {
	expirationTime := 24 * time.Hour
	finalTime := time.Now()

	type step struct {
		key string

		insertionTime time.Time
		shouldExpire  bool
	}

	// Each step represents a key that is inserted into the cache at a specific time.
	steps := []step{
		{
			key: "hello",

			insertionTime: finalTime.Add(-time.Minute),
			shouldExpire:  false,
		},
		{
			key: "foo",

			insertionTime: finalTime.Add(-(time.Hour * 24 * 2)),
			shouldExpire:  true,
		},
		{
			key: "bar",

			insertionTime: finalTime.Add(-(time.Hour * 25)),
			shouldExpire:  true,
		},
	}

	// Prepare the list of expected inserted and expired keys at the end of the test.

	var expectedInsertedKeys []string
	var expectedExpiredKeys []string

	for _, step := range steps {
		expectedInsertedKeys = append(expectedInsertedKeys, step.key)

		if step.shouldExpire {
			expectedExpiredKeys = append(expectedExpiredKeys, step.key)
		}
	}

	// Prepare spies to track the inserted and expired keys during the test.

	var actualInsertedKeys []string
	var actualExpiredKeys []string

	newEntryFunc := func(k string) int {
		actualInsertedKeys = append(actualInsertedKeys, k)
		return len(k)
	}

	expirationFunc := func(k string, v int) {
		actualExpiredKeys = append(actualExpiredKeys, k)
	}

	clock := &testClock{
		now: time.Now(), // will be set to the correct time during the test
	}

	options := []Option[string, int]{
		WithTTL[string, int](expirationTime),
		WithExpirationFunc[string, int](expirationFunc),
		withClock[string, int](clock),
	}

	cache := New(newEntryFunc, options...)

	// Insert the keys into the cache, advance the clock to the final time, then reap the cache.
	for _, step := range steps {
		clock.now = step.insertionTime
		cache.Get(step.key)
	}

	clock.now = finalTime
	cache.reap()

	// Validate that we inserted all the keys that we expected to insert.

	sort.Strings(expectedInsertedKeys)
	sort.Strings(actualInsertedKeys)
	if diff := cmp.Diff(expectedInsertedKeys, actualInsertedKeys); diff != "" {
		t.Fatalf("unexpected inserted keys (-want +got):\n%s", diff)
	}

	// Validate that we expired all the keys that we expected to expire, and no others.

	sort.Strings(expectedExpiredKeys)
	sort.Strings(actualExpiredKeys)

	if diff := cmp.Diff(expectedExpiredKeys, actualExpiredKeys); diff != "" {
		t.Fatalf("unexpected expired keys (-want +got):\n%s", diff)
	}
}

func TestGet_After_Reap(t *testing.T) {
	callCount := 0
	newEntryFunc := func(k string) int {
		callCount++
		return len(k)
	}

	clock := &testClock{
		now: time.Now(), // will be set to the correct time during the test
	}

	options := []Option[string, int]{
		WithTTL[string, int](time.Hour),
		withClock[string, int](clock),
	}

	cache := New(newEntryFunc, options...)

	// Insert a key into the cache.
	cache.Get("hello")

	// Advance the clock to the point where the key should expire.
	clock.now = clock.now.Add(time.Hour * 2)

	// Reap the cache.
	cache.reap()

	// Test that the cache returns the correct value for a key that has been added.
	value := cache.Get("hello")
	if value != 5 {
		t.Errorf("expected cache to return 5, got %d", value)
	}

	// Test that newEntryFunc was called again for the existing key.
	if callCount != 2 {
		t.Errorf("expected newEntryFunc to be called twice, got %d", callCount)
	}
}

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time {
	return c.now
}

var _ clock = &testClock{}
