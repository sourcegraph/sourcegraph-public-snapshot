package collections

import (
	"testing"

	"github.com/grafana/regexp"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)

	cmp := NaturalCompare[int]

	t.Run("Set can be created from another Set", func(t *testing.T) {
		c := NewSet(a.Values()...)
		sliceA, sliceC := a.Sorted(cmp), c.Sorted(cmp)
		require.Equal(t, sliceA, sliceC)
	})

	t.Run("Values returns all values of set", func(t *testing.T) {
		aVals, bVals := a.Sorted(cmp), b.Sorted(cmp)

		require.Equal(t, []int{1, 2, 3}, aVals)
		require.Equal(t, []int{2, 3, 4}, bVals)
		require.Equal(t, []int{}, NewSet[int]().Values())
	})

	t.Run("Has returns true if set contains the value", func(t *testing.T) {
		require.True(t, a.Has(1))
		require.True(t, a.Has(2))
		require.True(t, a.Has(3))
		require.False(t, a.Has(4))
	})

	t.Run("Add adds values to the set", func(t *testing.T) {
		s := NewSet(1)
		s.Add(2)
		require.True(t, s.Has(2))

		// multiple values can be added at once
		s.Add(3, 4)
		require.True(t, s.Has(3))
		require.True(t, s.Has(4))

		// adding nil values is a no-op
		s.Add()
		require.Equal(t, []int{1, 2, 3, 4}, s.Sorted(cmp))
	})

	t.Run("Remove removes values from the set", func(t *testing.T) {
		s := NewSet(1, 2, 3, 4)
		s.Remove(2)
		require.False(t, s.Has(2))

		// multiple values can be removed at once
		s.Remove(3, 4)
		require.False(t, s.Has(3))
		require.False(t, s.Has(4))

		// removing nil is a no-op
		s.Remove()
		require.Equal(t, []int{1}, s.Values())
	})

	t.Run("Contains returns true if set contains the other set", func(t *testing.T) {
		require.True(t, a.Contains(NewSet(1, 2)))
		require.True(t, a.Contains(NewSet(1, 2, 3)))
		require.False(t, a.Contains(b))

		// set always contains self
		require.True(t, a.Contains(a))

		// empty set is always contained
		require.True(t, a.Contains(NewSet[int]()))
	})

	t.Run("Union creates a new set with all values from both sets", func(t *testing.T) {
		union := Union(a, b).Sorted(cmp)
		require.Equal(t, []int{1, 2, 3, 4}, union)

		// order does not matter
		another := Union(b, a).Sorted(cmp)
		require.Equal(t, union, another)

		// union with self results in same set
		union = Union(a, a).Sorted(cmp)
		require.Equal(t, a.Sorted(cmp), union)
	})

	t.Run("Intersection creates a new set with values that are in both sets", func(t *testing.T) {
		intersection := Intersection(a, b).Sorted(cmp)
		require.Equal(t, []int{2, 3}, intersection)

		// intersection with self is the same set as self
		intersection = Intersection(a, a).Sorted(cmp)
		require.Equal(t, []int{1, 2, 3}, intersection)

		// intersection with empty set is empty set
		intersection = Intersection(a, NewSet[int]()).Sorted(cmp)
		require.Equal(t, []int{}, intersection)

		// intersection with set that has no common values is empty set
		intersection = Intersection(a, NewSet(4, 5, 6)).Sorted(cmp)
		require.Equal(t, []int{}, intersection)
	})

	t.Run("Difference returns values that are in current set but not the other", func(t *testing.T) {
		difference := a.Difference(b)
		require.Equal(t, []int{1}, difference.Values())

		// difference with self is empty set
		difference = a.Difference(a)
		require.Equal(t, []int{}, difference.Values())

		// difference with empty set is the same set
		difference = a.Difference(NewSet[int]())
		require.Equal(t, a.Sorted(cmp), difference.Sorted(cmp))
	})
	t.Run("String returns string representation", func(t *testing.T) {
		require.Regexp(t, regexp.MustCompile(`Set\[[1-3] [1-3] [1-3]]`), a)

		// empty set
		require.Equal(t, "Set[]", NewSet[int]().String())
	})
}
