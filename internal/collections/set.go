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

func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s Set[T]) Remove(values ...T) {
	for _, v := range values {
		delete(s, v)
	}
}

func (s Set[T]) Has(value T) bool {
	_, found := s[value]
	return found
}

// Values returns a slice with all the values in the set.
// The values are returned in an unspecified order.
func (s Set[T]) Values() []T {
	return maps.Keys(s)
}

// Sorted returns the values of the set in sorted order using the given
// comparator function.
//
// The comparator function should return true if the first argument is less than
// the second, and false otherwise.
//
// Example:
//
//	s.Sorted(func(a, b int) bool { return a < b })
func (s Set[T]) Sorted(comparator func(a, b T) bool) []T {
	vals := s.Values()
	sort.Slice(vals, func(i, j int) bool {
		return comparator(vals[i], vals[j])
	})
	return vals
}

// Difference returns a set with elements in s that are not in b.
func (s Set[T]) Difference(b Set[T]) Set[T] {
	diff := NewSet[T]()

	for v := range s {
		if !b.Has(v) {
			diff.Add(v)
		}
	}

	return diff
}

// Intersect returns a new set with elements that are in both s and b.
func (s Set[T]) Intersect(b Set[T]) Set[T] {
	return Intersection(s, b)
}

// Contains returns true if s has all the elements in b.
func (s Set[T]) Contains(b Set[T]) bool {
	// do not waste time on loop if b is bigger than s
	if len(b) > len(s) {
		return false
	}

	for v := range b {
		if !s.Has(v) {
			return false
		}
	}
	return true
}

// IsEmpty returns true if the set doesn't contain any elements.
func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

// Union returns a new set with all the elements from s and b
func (s Set[T]) Union(b Set[T]) Set[T] {
	return Union(s, b)
}

// String returns a string representation of the set.
func (s Set[T]) String() string {
	return fmt.Sprintf("Set%v", maps.Keys(s))
}

func getShortLong[T comparable](a, b Set[T]) (Set[T], Set[T]) {
	if len(a) < len(b) {
		return a, b
	}
	return b, a
}

// Union returns a new set with all the elements from a and b
func Union[T comparable](a, b Set[T]) Set[T] {
	short, long := getShortLong(a, b)
	union := NewSet(long.Values()...)

	union.Add(short.Values()...)
	return union
}

// Intersection returns a new set with all the elements that are in both a and b.
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
