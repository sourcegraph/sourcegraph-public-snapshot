package basestore

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CloseRows closes the given rows object. The resulting error is a multierror
// containing the error parameter along with any errors that occur during scanning
// or closing the rows object. The rows object is assumed to be non-nil.
//
// The signature of this function allows scan methods to be written uniformly:
//
//	func ScanThings(rows *sql.Rows, queryErr error) (_ []Thing, err error) {
//	    if queryErr != nil {
//	        return nil, queryErr
//	    }
//	    defer func() { err = CloseRows(rows, err) }()
//
//	    // read things from rows
//	}
//
// Scan methods should be called directly with the results of `*store.Query` to
// ensure that the rows are always properly handled.
//
//	things, err := ScanThings(store.Query(ctx, query))
func CloseRows(rows *sql.Rows, err error) error {
	return errors.Append(err, rows.Close(), rows.Err())
}

// NewSliceScanner returns a basestore scanner function that returns all
// the values of a query result. The given function is invoked multiple
// times with a SQL rows object to scan a single value.
func NewSliceScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) ([]T, error) {
	return func(rows *sql.Rows, queryErr error) (values []T, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			value, err := f(rows)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		return values, nil
	}
}

// CollectionReducer configures how scanners created by `NewCollectionReducerScanner` will
// group values belonging to the same map key.
type CollectionReducer[V, Vs any] interface {
	Create() Vs
	Reduce(collection Vs, value V) Vs
}

// SliceReducer can be used as a collection reducer for `NewCollectionReducerScanner` to
// collect values belonging to each key into a slice.
type SliceReducer[T any] struct{}

func (r SliceReducer[T]) Create() []T                        { return nil }
func (r SliceReducer[T]) Reduce(collection []T, value T) []T { return append(collection, value) }

// SingleValueReducer can be used as a collection reducer for `NewCollectionReducerScanner` to
// return the single value belonging to each key into a slice. If there are duplicates, the last
// value scanned will "win" for each key.
type SingleValueReducer[T any] struct{}

func (r SingleValueReducer[T]) Create() (_ T)                  { return }
func (r SingleValueReducer[T]) Reduce(collection T, value T) T { return value }

// NewKeyedCollectionScanner returns a basestore scanner function that returns the values of a
// query result organized as a map. The given function is invoked multiple times with a SQL rows
// object to scan a single map value. The given reducer provides a way to customize how multiple
// values are reduced into a collection.
func NewKeyedCollectionScanner[K comparable, V, Vs any](
	scanPair func(dbutil.Scanner) (K, V, error),
	reducer CollectionReducer[V, Vs],
) func(rows *sql.Rows, queryErr error) (map[K]Vs, error) {
	return func(rows *sql.Rows, queryErr error) (values map[K]Vs, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		values = map[K]Vs{}
		for rows.Next() {
			key, value, err := scanPair(rows)
			if err != nil {
				return nil, err
			}

			collection, ok := values[key]
			if !ok {
				collection = reducer.Create()
			}

			values[key] = reducer.Reduce(collection, value)
		}

		return values, nil
	}
}

// NewMapScanner returns a basestore scanner function that returns the values of a
// query result organized as a map. The given function is invoked multiple times with
// a SQL rows object to scan a single map value.
func NewMapScanner[K comparable, V any](f func(dbutil.Scanner) (K, V, error)) func(rows *sql.Rows, queryErr error) (map[K]V, error) {
	return NewKeyedCollectionScanner[K, V, V](f, SingleValueReducer[V]{})
}

// NewMapSliceScanner returns a basestore scanner function that returns the values
// of a query result organized as a map of slice values. The given function is invoked
// multiple times with a SQL rows object to scan a single map key value.
func NewMapSliceScanner[K comparable, V any](f func(dbutil.Scanner) (K, V, error)) func(rows *sql.Rows, queryErr error) (map[K][]V, error) {
	return NewKeyedCollectionScanner[K, V, []V](f, SliceReducer[V]{})
}

// NewSliceWithCountScanner returns a basestore scanner function that returns all
// the values of the query result, as well as the count from the `COUNT(*) OVER()`
// window function. The given function is invoked multiple times with a SQL rows
// object to scan a single value.
// Example query that would avail of this function, where we want only 10 rows but still
// the count of everything that would have been returned, without performing two separate queries:
// SELECT u.id, COUNT(*) OVER() as count FROM users LIMIT 10
func NewSliceWithCountScanner[T any](f func(dbutil.Scanner) (T, int, error)) func(rows *sql.Rows, queryErr error) ([]T, int, error) {
	return func(rows *sql.Rows, queryErr error) (values []T, totalCount int, err error) {
		if queryErr != nil {
			return nil, 0, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			value, count, err := f(rows)
			if err != nil {
				return nil, 0, err
			}

			totalCount = count
			values = append(values, value)
		}

		return values, totalCount, nil
	}
}

// NewFirstScanner returns a basestore scanner function that returns the
// first value of a query result (assuming there is at most one value).
// The given function is invoked with a SQL rows object to scan a single
// value.
func NewFirstScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) (T, bool, error) {
	return func(rows *sql.Rows, queryErr error) (value T, _ bool, err error) {
		if queryErr != nil {
			return value, false, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		if !rows.Next() {
			return value, false, nil
		}

		value, err = f(rows)
		return value, true, err
	}
}

// ScanAny scans a single T value from the given scanner.
func ScanAny[T any](s dbutil.Scanner) (value T, err error) {
	err = s.Scan(&value)
	return
}

// ScanNullString scans a single nullable string from the given scanner.
func ScanNullString(s dbutil.Scanner) (string, error) {
	var value sql.NullString
	if err := s.Scan(&value); err != nil {
		return "", err
	}

	return value.String, nil
}

// ScanNullInt64 scans a single int64 from the given scanner.
func ScanNullInt64(s dbutil.Scanner) (int64, error) {
	var value sql.NullInt64
	if err := s.Scan(&value); err != nil {
		return 0, err
	}

	return value.Int64, nil
}

// ScanInt32Array scans a single int32 array from the given scanner.
func ScanInt32Array(s dbutil.Scanner) ([]int32, error) {
	var value pq.Int32Array
	if err := s.Scan(&value); err != nil {
		return nil, err
	}

	return []int32(value), nil
}

var (
	ScanInt             = ScanAny[int]
	ScanStrings         = NewSliceScanner(ScanAny[string])
	ScanFirstString     = NewFirstScanner(ScanAny[string])
	ScanFirstNullString = NewFirstScanner(ScanNullString)
	ScanInts            = NewSliceScanner(ScanAny[int])
	ScanInt32s          = NewSliceScanner(ScanAny[int32])
	ScanInt64s          = NewSliceScanner(ScanAny[int64])
	Scanuint32s         = NewSliceScanner(ScanAny[uint32])
	ScanFirstInt        = NewFirstScanner(ScanAny[int])
	ScanFirstInt64      = NewFirstScanner(ScanAny[int64])
	ScanFirstNullInt64  = NewFirstScanner(ScanNullInt64)
	ScanFloats          = NewSliceScanner(ScanAny[float64])
	ScanFirstFloat      = NewFirstScanner(ScanAny[float64])
	ScanBools           = NewSliceScanner(ScanAny[bool])
	ScanFirstBool       = NewFirstScanner(ScanAny[bool])
	ScanTimes           = NewSliceScanner(ScanAny[time.Time])
	ScanFirstTime       = NewFirstScanner(ScanAny[time.Time])
	ScanNullTimes       = NewSliceScanner(ScanAny[*time.Time])
	ScanFirstNullTime   = NewFirstScanner(ScanAny[*time.Time])
	ScanFirstInt32Array = NewFirstScanner(ScanInt32Array)
)
