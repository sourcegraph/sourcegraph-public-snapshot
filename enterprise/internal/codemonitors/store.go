package codemonitors

import (
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// Store exposes methods to read and write campaigns domain models
// from persistent storage.
type Store struct {
	*basestore.Store
	now func() time.Time
}

// NewStore returns a new Store backed by the given db.
func NewStore(db dbutil.DB) *Store {
	return NewStoreWithClock(db, timeutil.Now)
}

// NewStoreWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewStoreWithClock(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{Store: basestore.NewWithDB(db, sql.TxOptions{}), now: clock}
}

// With creates a new store with the underlying database handle from the given store.
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store: s.Store.With(other),
		now:   timeutil.Now,
	}
}

// Clock returns the clock of the underlying store.
func (s *Store) Clock() func() time.Time {
	return s.now
}
