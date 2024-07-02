package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestOrderedSet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := rapid.SliceOfN(rapid.IntRange(-3, 6), 0, 10).Draw(t, "data")
		set := NewSet(data...)
		uniquedData := set.Values()
		ordset := NewOrderedSet(uniquedData...)
		require.Equal(t, uniquedData, ordset.Values())

		otherData := rapid.SliceOfN(rapid.IntRange(-5, 5), 10, 10).Draw(t, "data")
		for _, x := range otherData {
			require.Equal(t, set.Has(x), ordset.Has(x))
		}

		for _, x := range uniquedData {
			require.True(t, ordset.Has(x))
			ordset.Remove(x)
			require.False(t, ordset.Has(x))
		}
	})
}
