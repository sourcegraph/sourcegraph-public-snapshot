package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// Interface is the interface describing a code insights store. See the Store struct
// for actual API usage.
type Interface interface {
	SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error)
}

var _ Interface = &Store{}

// Store exposes methods to read and write code insights domain models from
// persistent storage.
type Store struct {
	*basestore.Store
	now func() time.Time
}

// New returns a new Store backed by the given Timescale db.
func New(db dbutil.DB) *Store {
	return NewWithClock(db, timeutil.Now)
}

// NewWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewWithClock(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{Store: basestore.NewWithDB(db, sql.TxOptions{}), now: clock}
}

var _ basestore.ShareableStore = &Store{}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *Store) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other), now: s.now}
}

// SeriesPoint describes a single insights' series data point.
type SeriesPoint struct {
	Time  time.Time
	Value float64
}

// SeriesPointsOpts describes options for querying insights' series data points.
type SeriesPointsOpts struct {
	// SeriesID is the unique series ID to query, if non-nil.
	SeriesID *int32

	// Time ranges to query from/to, if non-nil.
	From, To *time.Time

	// Limit is the number of data points to query, if non-zero.
	Limit int
}

// SeriesPoints queries data points over time for a specific insights' series.
func (s *Store) SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error) {
	return nil, errors.New("not yet implemented")
}
