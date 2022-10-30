package basestore

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

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

// CollectionReducer configures how scanners created by `NewKeyedCollectionScanner` will
// group values belonging to the same map key.
type CollectionReducer[V, Vs any] interface {
	Create() Vs
	Reduce(collection Vs, value V) Vs
}

// SliceReducer can be used as a collection reducer for `NewKeyedCollectionScanner` to
// collect values belonging to each key into a slice.
type SliceReducer[T any] struct{}

func (r SliceReducer[T]) Create() []T                        { return nil }
func (r SliceReducer[T]) Reduce(collection []T, value T) []T { return append(collection, value) }

// SingleValueReducer can be used as a collection reducer for `NewKeyedCollectionScanner` to
// return the single value belonging to each key into a slice. If there are duplicates, the last
// value scanned will "win" for each key.
type SingleValueReducer[T any] struct{}

func (r SingleValueReducer[T]) Create() (_ T)                  { return }
func (r SingleValueReducer[T]) Reduce(collection T, value T) T { return value }
