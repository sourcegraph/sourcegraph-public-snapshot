package database

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDatabaseCacheEvictionWhileHeld(t *testing.T) {
	// keep track of what db mocks are closed
	closed := map[IDatabase]bool{}
	// protected concurrent access to closed map
	var mutex sync.Mutex

	// openTestDatabase creates a new IDatabase that inserts an entry
	// for itself into the close map when its Close method is called.
	openTestDatabase := func() (IDatabase, error) {
		db := NewMockIDatabase()
		db.CloseFunc.SetDefaultHook(func() error {
			mutex.Lock()
			defer mutex.Unlock()
			closed[db] = true
			return nil
		})

		return db, nil
	}

	// isOpen returns true if the database has not yet been closed.
	isOpen := func(db IDatabase) bool {
		mutex.Lock()
		defer mutex.Unlock()
		return !closed[db]
	}

	// isOpenFor100ms will call isOpen for the given database until it
	// has closed or 250ms has elapsed. This is used to test whether or
	// not a database handle has been closed by an LRU eviction.
	isOpenFor250ms := func(db IDatabase) bool {
		for i := 0; i < 100; i++ {
			if isOpen(db) {
				time.Sleep(time.Millisecond)
				continue
			}

			return false
		}

		return true
	}

	cache, err := NewDatabaseCache(2)
	if err != nil {
		t.Fatalf("unexpected error creating database cache: %s", err)
	}

	// reference to a db handle that outlives the cache entry
	var dbRef IDatabase

	// cache: foo
	if err := cache.WithDatabase("foo", openTestDatabase, func(db1 IDatabase) error {
		dbRef = db1

	outer:
		for {
			// cache: bar,foo
			if err := cache.WithDatabase("bar", openTestDatabase, func(_ IDatabase) error {
				return nil
			}); err != nil {
				return err
			}

			// cache: baz, bar
			// expected: foo was evicted but should not be closed
			// possible: another key was evicted instead due to ristretto's heuristic counters
			if err := cache.WithDatabase("baz", openTestDatabase, func(_ IDatabase) error {
				return nil
			}); err != nil {
				return err
			}

			// In order to get around the possibility above, we ensure that the key foo
			// is evicted from the cache before moving on. In the event that either the
			// key bar or baz was evicted instead, we retry from the outer loop.

			for {
				_, ok1 := cache.cache.Get("foo")
				_, ok2 := cache.cache.Get("bar")
				_, ok3 := cache.cache.Get("baz")

				if !ok1 {
					break outer
				}

				if !ok2 || !ok3 {
					continue outer
				}

				// Nothing may be evicted yet as ristretto is "eventually consistent",
				// so we need to loop here as well until _something_ has been evicted.
			}
		}

		// cache: foo, bar
		// note: this version of foo should be a fresh connection
		return cache.WithDatabase("foo", openTestDatabase, func(db2 IDatabase) error {
			if db1 == db2 {
				return errors.New("unexpected cached database")
			}

			// evicted database stays open while held
			if !isOpenFor250ms(db1) {
				return fmt.Errorf("db1 unexpectedly closed")
			}
			if !isOpen(db2) {
				return fmt.Errorf("db2 unexpectedly closed")
			}

			return nil
		})
	}); err != nil {
		t.Fatalf("unexpected error during test: %s", err)
	}

	// evicted database is eventually closed
	if isOpenFor250ms(dbRef) {
		t.Fatalf("database remained unexpectedly open")
	}
}
