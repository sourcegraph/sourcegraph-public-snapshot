package basestore

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// NewCallbackScanner returns a basestore scanner function that invokes the given
// function on every SQL row object in the given query result set. If the callback
// function returns a false-valued flag, the remaining rows are discarded.
func NewCallbackScanner(f func(dbutil.Scanner) (bool, error)) func(rows Rows, queryErr error) error {
	return func(rows Rows, queryErr error) (err error) {
		if queryErr != nil {
			return queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			if ok, err := f(rows); err != nil {
				return err
			} else if !ok {
				break
			}
		}

		return nil
	}
}

// NewFirstScanner returns a basestore scanner function that returns the
// first value of a query result (assuming there is at most one value).
// The given function is invoked with a SQL rows object to scan a single
// value.
func NewFirstScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows Rows, queryErr error) (T, bool, error) {
	return func(rows Rows, queryErr error) (value T, called bool, _ error) {
		scanner := func(s dbutil.Scanner) (_ bool, err error) {
			called = true
			value, err = f(s)
			return false, err
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return value, called, err
	}
}

// NewSliceScanner returns a basestore scanner function that returns all
// the values of a query result. The given function is invoked multiple
// times with a SQL rows object to scan a single value.
func NewSliceScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows Rows, queryErr error) ([]T, error) {
	return NewFilteredSliceScanner(func(s dbutil.Scanner) (T, bool, error) {
		value, err := f(s)
		return value, true, err
	})
}

// NewFilteredSliceScanner returns a basestore scanner function that returns
// filtered values  of a query result. The given function is invoked multiple
// times with a SQL rows object to scan a single value. If the boolean flag
// returned by the function is false, the associated value is not added to the
// returned slice.
func NewFilteredSliceScanner[T any](f func(dbutil.Scanner) (T, bool, error)) func(rows Rows, queryErr error) ([]T, error) {
	return func(rows Rows, queryErr error) (values []T, _ error) {
		scanner := func(s dbutil.Scanner) (bool, error) {
			value, ok, err := f(s)
			if err != nil {
				return false, err
			}
			if ok {
				values = append(values, value)
			}

			return true, nil
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
func NewSliceWithCountScanner[T any](f func(dbutil.Scanner) (T, int, error)) func(rows Rows, queryErr error) ([]T, int, error) {
	return func(rows Rows, queryErr error) (values []T, totalCount int, _ error) {
		scanner := func(s dbutil.Scanner) (bool, error) {
			value, count, err := f(s)
			if err != nil {
				return false, err
			}

			totalCount = count
			values = append(values, value)
			return true, nil
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return values, totalCount, err
	}
}

// NewKeyedCollectionScanner returns a basestore scanner function that returns the values of a
// query result organized as a map. The given function is invoked multiple times with a SQL rows
// object to scan a single map value. The given reducer provides a way to customize how multiple
// values are reduced into a collection.
func NewKeyedCollectionScanner[K comparable, V, Vs any, Map keyedMap[K, Vs]](
	values Map,
	scanPair func(dbutil.Scanner) (K, V, error),
	reducer CollectionReducer[V, Vs],
) func(rows Rows, queryErr error) error {
	return func(rows Rows, queryErr error) error {
		scanner := func(s dbutil.Scanner) (bool, error) {
			key, value, err := scanPair(s)
			if err != nil {
				return false, err
			}

			collection, ok := values.Get(key)
			if !ok {
				collection = reducer.Create()
			}

			values.Set(key, reducer.Reduce(collection, value))
			return true, nil
		}

		err := NewCallbackScanner(scanner)(rows, queryErr)
		return err
	}
}

// NewMapScanner returns a basestore scanner function that returns the values of a
// query result organized as a map. The given function is invoked multiple times with
// a SQL rows object to scan a single map value.
func NewMapScanner[K comparable, V any](f func(dbutil.Scanner) (K, V, error)) func(rows Rows, queryErr error) (map[K]V, error) {
	return func(rows Rows, queryErr error) (map[K]V, error) {
		m := NewUnorderedmap[K, V]()
		err := NewKeyedCollectionScanner[K, V, V](m, f, SingleValueReducer[V]{})(rows, queryErr)
		return m.ToMap(), err
	}
}

// NewMapSliceScanner returns a basestore scanner function that returns the values
// of a query result organized as a map of slice values. The given function is invoked
// multiple times with a SQL rows object to scan a single map key value.
func NewMapSliceScanner[K comparable, V any](f func(dbutil.Scanner) (K, V, error)) func(rows Rows, queryErr error) (map[K][]V, error) {
	return func(rows Rows, queryErr error) (map[K][]V, error) {
		m := NewUnorderedmap[K, []V]()
		err := NewKeyedCollectionScanner[K, V, []V](m, f, SliceReducer[V]{})(rows, queryErr)
		return m.ToMap(), err
	}
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

type keyedMap[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
	Len() int
	Values() []V
	ToMap() map[K]V
}

type UnorderedMap[K comparable, V any] struct {
	m map[K]V
}

func NewUnorderedmap[K comparable, V any]() *UnorderedMap[K, V] {
	return &UnorderedMap[K, V]{m: make(map[K]V)}
}

func (m UnorderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.m[key]
	return v, ok
}

func (m UnorderedMap[K, V]) Set(key K, val V) {
	m.m[key] = val
}

func (m UnorderedMap[K, V]) Len() int {
	return len(m.m)
}

func (m UnorderedMap[K, V]) Values() []V {
	return maps.Values(m.m)
}

func (m *UnorderedMap[K, V]) ToMap() map[K]V {
	return m.m
}

type OrderedMap[K comparable, V any] struct {
	m *orderedmap.OrderedMap[K, V]
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: orderedmap.New[K, V]()}
}

func (m OrderedMap[K, V]) Get(key K) (V, bool) {
	return m.m.Get(key)
}

func (m OrderedMap[K, V]) Set(key K, val V) {
	m.m.Set(key, val)
}

func (m OrderedMap[K, V]) Len() int {
	return m.m.Len()
}

func (m OrderedMap[K, V]) Values() []V {
	values := make([]V, 0, m.m.Len())
	for pair := m.m.Oldest(); pair != nil; pair = pair.Next() {
		values = append(values, pair.Value)
	}
	return values
}

func (m *OrderedMap[K, V]) ToMap() map[K]V {
	ret := make(map[K]V, m.m.Len())
	for pair := m.m.Oldest(); pair != nil; pair = pair.Next() {
		ret[pair.Key] = pair.Value
	}
	return ret
}
