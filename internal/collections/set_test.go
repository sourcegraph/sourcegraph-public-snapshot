package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)

	comparator := func(a, b int) bool {
		return a < b
	}

	t.Run("Set can be created from another Set", func(t *testing.T) {
		c := NewSet(a.Values()...)
		sliceA, sliceC := a.SortedValues(comparator), c.SortedValues(comparator)
		require.Equal(t, sliceA, sliceC)
	})

	t.Run("Values returns all values of set", func(t *testing.T) {
		aVals, bVals := a.SortedValues(comparator), b.SortedValues(comparator)

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
		union := Union(a, b)
		unionVals := union.SortedValues(comparator)
		require.Equal(t, []int{1, 2, 3, 4}, unionVals)
	})

	t.Run("Intersection creates a new set with values that are in both sets", func(t *testing.T) {
		itrsc := Intersection(a, b)
		itrscVals := itrsc.SortedValues(comparator)
		require.Equal(t, []int{2, 3}, itrscVals)
	})

	t.Run("Difference returns values that are in current set but not the other", func(t *testing.T) {
		itrsc := a.Difference(b)
		require.Equal(t, []int{1}, itrsc.Values())
	})
}
