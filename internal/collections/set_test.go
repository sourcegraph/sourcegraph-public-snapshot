package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)

	cmp := func(a, b int) bool {
		return a < b
	}

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

	t.Run("Contains returns true if set contains value", func(t *testing.T) {
		require.True(t, a.Contains(1))
		require.True(t, a.Contains(2))
		require.True(t, a.Contains(3))
		require.False(t, a.Contains(4))
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
		itrsc := Intersection(a, b).Sorted(cmp)
		require.Equal(t, []int{2, 3}, itrsc)

		// intersection with self is the same set as self
		itrsc = Intersection(a, a).Sorted(cmp)
		require.Equal(t, []int{1, 2, 3}, itrsc)

		// intersection with empty set is empty set
		itrsc = Intersection(a, NewSet[int]()).Sorted(cmp)
		require.Equal(t, []int{}, itrsc)

		// intersection with set that has no common values is empty set
		itrsc = Intersection(a, NewSet(4, 5, 6)).Sorted(cmp)
		require.Equal(t, []int{}, itrsc)
	})

	t.Run("Difference returns values that are in current set but not the other", func(t *testing.T) {
		itrsc := a.Difference(b)
		require.Equal(t, []int{1}, itrsc.Values())

		// difference with self is empty set
		itrsc = a.Difference(a)
		require.Equal(t, []int{}, itrsc.Values())

		// difference with empty set is the same set
		itrsc = a.Difference(NewSet[int]())
		require.Equal(t, a.Sorted(cmp), itrsc.Sorted(cmp))
	})
}
