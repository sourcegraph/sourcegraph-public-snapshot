package database

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDatabaseCacheEvictionWhileHeld(t *testing.T) {
	cache, err := NewDatabaseCache(2)
	if err != nil {
		t.Fatalf("unexpected error creating database cache: %s", err)
	}

	// reference to a db handle that outlives the cache entry
	var dbRef *Database

	// cache: foo
	if err := cache.WithDatabase("foo", openTestDatabase, func(db1 *Database) error {
		dbRef = db1

	outer:
		for {
			// cache: bar,foo
			if err := cache.WithDatabase("bar", openTestDatabase, func(_ *Database) error {
				return nil
			}); err != nil {
				return err
			}

			// cache: baz, bar
			// expected: foo was evicted but should not be closed
			// possible: another key was evicted instead due to ristretto's heuristic counters
			if err := cache.WithDatabase("baz", openTestDatabase, func(_ *Database) error {
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
		return cache.WithDatabase("foo", openTestDatabase, func(db2 *Database) error {
			if db1 == db2 {
				return errors.New("unexpected cached database")
			}

			// evicted database stays open while held
			_ = readMetaLoop(db1)
			meta1, err1 := ReadMeta(db1.db)
			meta2, err2 := ReadMeta(db2.db)

			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			if meta1.LSIFVersion != "0.4.3" || meta2.LSIFVersion != "0.4.3" {
				return fmt.Errorf("unexpected lsif versions: want=%q have=%q and %q", "0.4.3", meta1.LSIFVersion, meta2.LSIFVersion)
			}

			return nil
		})
	}); err != nil {
		t.Fatalf("unexpected error during test: %s", err)
	}

	// evicted database is eventually closed
	if err := readMetaLoop(dbRef); err == nil {
		t.Fatalf("unexpected nil error")
	} else if !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("unexpected error: want=%q have=%q", "database is closed", err)
	}
}

// readMetaLoop attempts to read the metadata from the given database and returns the
// first error that occurs. The ReadMeta function is re-invoked until an error occurs
// or until there were 100 attempts. This function is used to determine if the database
// has been closed (non-nil error), or to ensure that the database remains open for at
// least 100ms (nil-valued error).
func readMetaLoop(db *Database) (err error) {
	for i := 0; i < 100; i++ {
		if _, err = ReadMeta(db.db); err != nil {
			break
		}

		time.Sleep(time.Millisecond)
	}

	return err
}
