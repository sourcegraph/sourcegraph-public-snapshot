package collections

import (
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
	for _, v := range values {
		s.Add(v)
	}
	return s
}

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Remove(value T) {
	delete(s, value)
}

func (s Set[T]) Contains(value T) bool {
	_, found := s[value]
	return found
}

// Values returns a slice containing all the values in the set.
// The values are returned in an unspecified order.
func (s Set[T]) Values() []T {
	return maps.Keys(s)
}

// SortedValues returns the values of the set in sorted order using the given
// comparator function.
//
// The comparator function should return true if the first argument is less than
// the second, and false otherwise.
//
// Example:
//
//	s.SortedValues(func(a, b int) bool { return a < b })
func (s Set[T]) SortedValues(comparator func(a, b T) bool) []T {
	vals := s.Values()
	sort.Slice(vals, func(i, j int) bool {
		return comparator(vals[i], vals[j])
	})
	return vals
}

// Difference returns a set containing the elements in s that are not in b.
func (s Set[T]) Difference(b Set[T]) Set[T] {
	diff := NewSet[T]()

	for v := range s {
		if !b.Contains(v) {
			diff.Add(v)
		}
	}

	return diff
}

// Intersect returns a new set containing the elements that are in both s and b.
func (s Set[T]) Intersect(b Set[T]) Set[T] {
	return Intersection(s, b)
}

// Union returns a new set with all the elements from s and b
func (s Set[T]) Union(b Set[T]) Set[T] {
	return Union(s, b)
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

	for v := range short {
		union.Add(v)
	}
	return union
}

// Intersection returns a new set containing the elements that are in both a and b.
func Intersection[T comparable](a, b Set[T]) Set[T] {
	itrsc := NewSet[T]()
	short, long := getShortLong(a, b)

	for v := range short {
		if long.Contains(v) {
			itrsc.Add(v)
		}
	}

	return itrsc
}
