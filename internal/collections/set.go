package collections

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

// Set is a set (collection of unique elements) implemented as a map.
// T must be a comparable type (implementing sort.Interface or == operator).
// The zero value for Set is nil, so it needs to be initialized as Set[T]{}
// or with NewSet[T]().
type Set[T comparable] map[T]struct{}

// NewSet creates a Set[T] with the given values.
// T must be a comparable type (implementing sort.Interface or == operator).
//
// Example:
//
//	s := NewSet[int](1, 2, 3)
func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{}
	s.Add(values...)
	return s
}

func (a Set[T]) Add(values ...T) {
	for _, v := range values {
		a[v] = struct{}{}
	}
}

func (a Set[T]) Remove(values ...T) {
	for _, v := range values {
		delete(a, v)
	}
}

func (a Set[T]) Has(value T) bool {
	_, found := a[value]
	return found
}

// Values returns a slice with all the values in the set.
// The values are returned in an unspecified order.
func (a Set[T]) Values() []T {
	return maps.Keys(a)
}

// Sorted returns the values of the set in sorted order using the given
// comparator function.
//
// The comparator function should return true if the first argument is less than
// the second, and false otherwise.
//
// Example:
//
//	a.Sorted(func(a, b int) bool { return a < b })
func (a Set[T]) Sorted(comparator func(a, b T) bool) []T {
	vals := a.Values()
	sort.Slice(vals, func(i, j int) bool {
		return comparator(vals[i], vals[j])
	})
	return vals
}

// Difference returns a set with elements in A that are not in B.
func (a Set[T]) Difference(b Set[T]) Set[T] {
	diff := NewSet[T]()

	for v := range a {
		if !b.Has(v) {
			diff.Add(v)
		}
	}

	return diff
}

// Intersect returns a new set with elements that are in both A and B.
func (a Set[T]) Intersect(b Set[T]) Set[T] {
	return Intersection(a, b)
}

// Contains returns true if A has all the elements in B.
func (a Set[T]) Contains(b Set[T]) bool {
	// Do not waste time on loop if B is bigger than A.
	if len(b) > len(a) {
		return false
	}

	for v := range b {
		if !a.Has(v) {
			return false
		}
	}
	return true
}

// IsEmpty returns true if the set doesn't contain any elements.
func (a Set[T]) IsEmpty() bool {
	return len(a) == 0
}

// Union returns a new set with all the elements from A and B.
func (a Set[T]) Union(b Set[T]) Set[T] {
	return Union(a, b)
}

// String returns a string representation of the set.
func (a Set[T]) String() string {
	return fmt.Sprintf("Set%v", maps.Keys(a))
}

func getShortLong[T comparable](a, b Set[T]) (Set[T], Set[T]) {
	if len(a) < len(b) {
		return a, b
	}
	return b, a
}

// Union returns a new set with all the elements from A and B.
func Union[T comparable](a, b Set[T]) Set[T] {
	short, long := getShortLong(a, b)
	union := NewSet(long.Values()...)

	union.Add(short.Values()...)
	return union
}

// Intersection returns a new set with all the elements that are in both A and B.
func Intersection[T comparable](a, b Set[T]) Set[T] {
	itrsc := NewSet[T]()
	short, long := getShortLong(a, b)

	for v := range short {
		if long.Has(v) {
			itrsc.Add(v)
		}
	}

	return itrsc
}
