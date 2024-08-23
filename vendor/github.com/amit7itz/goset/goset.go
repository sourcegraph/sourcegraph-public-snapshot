package goset

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Set represents a set data structure.
// You should not call it directly, use NewSet() or FromSlice()
type Set[T comparable] struct {
	store map[T]struct{}
}

// NewSet returns a new Set of the given items
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{store: make(map[T]struct{}, 0)}
	set.Add(items...)
	return set
}

// FromSlice returns a new Set with all the items of the slice.
func FromSlice[T comparable](slice []T) *Set[T] {
	set := NewSet[T]()
	set.Add(slice...)
	return set
}

// String returns a string that represents the Set
func (s *Set[T]) String() string {
	var t T
	str := fmt.Sprintf("Set[%s]{", reflect.TypeOf(t).String())
	itemsStr := make([]string, 0, s.Len())
	for item := range s.store {
		itemsStr = append(itemsStr, fmt.Sprintf("%v", item))
	}
	str += strings.Join(itemsStr, " ")
	str += "}"
	return str
}

// Add adds item(s) to the Set
func (s *Set[T]) Add(items ...T) {
	for _, item := range items {
		s.store[item] = struct{}{}
	}
}

// Remove removes a single item from the Set. Returns error if the item is not in the Set
// See also: Discard()
func (s *Set[T]) Remove(item T) error {
	if s.Contains(item) {
		delete(s.store, item)
		return nil
	}
	return fmt.Errorf("item not found: %v ", item)
}

// Discard removes item(s) from the Set if exist
// See also: Remove()
func (s *Set[T]) Discard(items ...T) {
	for _, item := range items {
		delete(s.store, item)
	}
}

// Len returns the number of items in the Set
func (s *Set[T]) Len() int {
	return len(s.store)
}

// IsEmpty returns true if there are no items in the Set
func (s *Set[T]) IsEmpty() bool {
	return len(s.store) == 0
}

// Contains returns whether an item is in the Set
func (s *Set[T]) Contains(item T) bool {
	_, ok := s.store[item]
	return ok
}

// Update adds all the items from the other Sets to the current Set
func (s *Set[T]) Update(others ...*Set[T]) {
	for _, other := range others {
		for item := range other.store {
			s.Add(item)
		}
	}
}

// Pop removes an arbitrary item from the Set and returns it. Returns error if the Set is empty
func (s *Set[T]) Pop() (T, error) {
	var item T
	if s.IsEmpty() {
		return item, errors.New("set is empty")
	}
	for item = range s.store {
		break
	}
	s.Discard(item)
	return item, nil
}

// Copy returns a new Set with the same items as the current Set
func (s *Set[T]) Copy() *Set[T] {
	set := NewSet[T]()
	for item := range s.store {
		set.Add(item)
	}
	return set
}

// Items returns a slice of all the Set items
func (s *Set[T]) Items() []T {
	items := make([]T, 0, s.Len())
	for item := range s.store {
		items = append(items, item)
	}
	return items
}

// Equal returns whether the current Set contains the same items as the other one
func (s *Set[T]) Equal(other *Set[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	for item := range s.store {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// Union returns a new Set of the items from the current set and all others
func (s *Set[T]) Union(others ...*Set[T]) *Set[T] {
	unionSet := s.Copy()
	unionSet.Update(others...)
	return unionSet
}

// Intersection returns a new Set with the common items of the current set and all others.
func (s *Set[T]) Intersection(others ...*Set[T]) *Set[T] {
	intersectionSet := NewSet[T]()
	for item := range s.store {
		inAllOthers := true
		for _, other := range others {
			if !other.Contains(item) {
				inAllOthers = false
				break
			}
		}
		if inAllOthers {
			intersectionSet.Add(item)
		}
	}
	return intersectionSet
}

// Difference returns a new Set of all the items in the current Set that are not in any of the others
func (s *Set[T]) Difference(others ...*Set[T]) *Set[T] {
	differenceSet := NewSet[T]()
	for item := range s.store {
		inAnyOther := false
		for _, other := range others {
			if other.Contains(item) {
				inAnyOther = true
				break
			}
		}
		if !inAnyOther {
			differenceSet.Add(item)
		}
	}
	return differenceSet
}

// SymmetricDifference returns all the items that exist in only one of the Sets
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	symmetricDifferenceSet := NewSet[T]()
	for item := range s.store {
		if !other.Contains(item) {
			symmetricDifferenceSet.Add(item)
		}
	}
	for item := range other.store {
		if !s.Contains(item) {
			symmetricDifferenceSet.Add(item)
		}
	}
	return symmetricDifferenceSet
}

// IsDisjoint returns whether the two Sets have no item in common
func (s *Set[T]) IsDisjoint(other *Set[T]) bool {
	intersection := s.Intersection(other)
	return intersection.IsEmpty()
}

// IsSubset returns whether all the items of the current set exist in the other one
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	intersection := s.Intersection(other)
	return intersection.Len() == s.Len()
}

// IsSuperset returns whether all the items of the other set exist in the current one
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}
