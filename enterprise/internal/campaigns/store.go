package campaigns

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// seededRand is used to populate the RandID fields on CampaignSpec and
// ChangesetSpec when creating them.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// ErrNoResults is returned by Store method calls that found no results.
var ErrNoResults = errors.New("no results")

// Store exposes methods to read and write campaigns domain models
// from persistent storage.
type Store struct {
	*basestore.Store
	now func() time.Time
}

// NewStore returns a new Store backed by the given db.
func NewStore(db dbutil.DB) *Store {
	return NewStoreWithClock(db, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewStoreWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewStoreWithClock(db dbutil.DB, clock func() time.Time) *Store {
	handle := basestore.NewHandleWithDB(db)
	return &Store{Store: basestore.NewWithHandle(handle), now: clock}
}

// Clock returns the clock used by the Store.
func (s *Store) Clock() func() time.Time { return s.now }

// DB returns the underlying dbutil.DB that this Store was
// instantiated with.
// It's here for legacy reason to pass the dbutil.DB to a repos.Store while
// repos.Store doesn't accept a basestore.TransactableHandle yet.
func (s *Store) DB() dbutil.DB { return s.Handle().DB() }

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

// Transact creates a new transaction.
// It's required to implement this method and wrap the Transact method of the
// underlying basestore.Store.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{Store: txBase, now: s.now}, nil
}

var NoTransactionError = errors.New("Not in a transaction")

var lockNamespace = int32(fnv1.HashString32("campaigns"))

// TryAcquireAdvisoryLock will attempt to acquire an advisory lock using key
// and is non blocking. If a lock is acquired, "true, nil" will be returned.
// It must be called from within a transaction or "false, NoTransactionError" is returned
func (s *Store) TryAcquireAdvisoryLock(ctx context.Context, key string) (bool, error) {
	if ok := s.Store.InTransaction(); !ok {
		return false, NoTransactionError
	}

	q := lockQuery(key)
	ok, _, err := basestore.ScanFirstBool(s.Store.Query(ctx, q))
	if err != nil || !ok {
		return false, err
	}
	return true, nil
}

func lockQuery(key string) *sqlf.Query {
	// Postgres advisory lock ids are a global namespace within one database.
	// It's very unlikely that another part of our application uses a lock
	// namespace identically to this one. It's equally unlikely that there are
	// lock id conflicts for different permissions, but if it'd happen, no safety
	// guarantees would be violated, since those two different users would simply
	// have to wait on the other's update to finish, using stale permissions until
	// it would.
	lockID := int32(fnv1.HashString32(key))
	return sqlf.Sprintf(
		lockQueryFmtStr,
		lockNamespace,
		lockID,
	)
}

const lockQueryFmtStr = `
-- source: enterprise/internal/campaigns/store/store.go:TryAcquireAdvisoryLock
SELECT pg_try_advisory_xact_lock(%s, %s)
`

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

func (s *Store) queryCount(ctx context.Context, q *sqlf.Query) (int, error) {
	count, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return count, err
	}
	return count, nil
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (err error)

func scanAll(rows *sql.Rows, scan scanFunc) (err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err = scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func jsonbColumn(metadata interface{}) (msg json.RawMessage, err error) {
	switch m := metadata.(type) {
	case nil:
		msg = json.RawMessage("{}")
	case string:
		msg = json.RawMessage(m)
	case []byte:
		msg = m
	case json.RawMessage:
		msg = m
	default:
		msg, err = json.MarshalIndent(m, "        ", "    ")
	}
	return
}

func jsonSetColumn(ids []int64) ([]byte, error) {
	set := make(map[int64]*struct{}, len(ids))
	for _, id := range ids {
		set[id] = nil
	}
	return json.Marshal(set)
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullInt64Column(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
