package basestore

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// NewCallbackScanner returns a basestore scanner function that invokes
// the given function on every SQL row object in the given query result
// set.
func NewCallbackScanner(f func(dbutil.Scanner) error) func(rows *sql.Rows, queryErr error) error {
	return func(rows *sql.Rows, queryErr error) (err error) {
		if queryErr != nil {
			return queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			if err := f(rows); err != nil {
				return err
			}
		}

		return nil
	}
}

// NewFirstScanner returns a basestore scanner function that returns the
// first value of a query result (assuming there is at most one value).
// The given function is invoked with a SQL rows object to scan a single
// value.
func NewFirstScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) (T, bool, error) {
	return func(rows *sql.Rows, queryErr error) (value T, called bool, _ error) {
		scanner := func(s dbutil.Scanner) (err error) {
			called = true
			value, err = f(s)
			return err
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return value, called, err
	}
}

// NewSliceScanner returns a basestore scanner function that returns all
// the values of a query result. The given function is invoked multiple
// times with a SQL rows object to scan a single value.
func NewSliceScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) ([]T, error) {
	return func(rows *sql.Rows, queryErr error) (values []T, _ error) {
		scanner := func(s dbutil.Scanner) error {
			value, err := f(s)
			if err != nil {
				return err
			}

			values = append(values, value)
			return nil
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return values, err
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
	return func(rows *sql.Rows, queryErr error) (values []T, totalCount int, _ error) {
		scanner := func(s dbutil.Scanner) error {
			value, count, err := f(s)
			if err != nil {
				return err
			}

			totalCount = count
			values = append(values, value)
			return nil
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return values, totalCount, err
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
	return func(rows *sql.Rows, queryErr error) (map[K]Vs, error) {
		values := map[K]Vs{}
		scanner := func(s dbutil.Scanner) error {
			key, value, err := scanPair(s)
			if err != nil {
				return err
			}

			collection, ok := values[key]
			if !ok {
				collection = reducer.Create()
			}

			values[key] = reducer.Reduce(collection, value)
			return nil
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return values, err
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
