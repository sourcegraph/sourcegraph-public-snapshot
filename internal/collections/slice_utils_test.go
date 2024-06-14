package collections

import (
	stdcmp "cmp"
	"pgregory.net/rapid"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func Test_SplitIntoChunks(t *testing.T) {
	t.Run("Splits a slice into chunks of size 3", func(t *testing.T) {
		got, err := SplitIntoChunks([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3)
		require.NoError(t, err)
		want := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Type of slice does not matter", func(t *testing.T) {
		got, err := SplitIntoChunks([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, 4)
		require.NoError(t, err)
		want := [][]string{{"a", "b", "c", "d"}, {"e", "f", "g", "h"}, {"i", "j"}}
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Splits into 1 chunk if slice is smaller than requested chunk size", func(t *testing.T) {
		got, err := SplitIntoChunks([]int{1, 2, 3}, 4)
		require.NoError(t, err)
		want := [][]int{{1, 2, 3}}
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Works with chunk size of 1", func(t *testing.T) {
		got, err := SplitIntoChunks([]int{1, 2, 3}, 1)
		require.NoError(t, err)
		want := [][]int{{1}, {2}, {3}}
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Works with empty slice", func(t *testing.T) {
		got, err := SplitIntoChunks([]int{}, 4)
		require.NoError(t, err)
		want := make([][]int, 0)
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("returns error for chunk size of 0", func(t *testing.T) {
		_, err := SplitIntoChunks([]int{1, 2, 3}, 0)
		require.Error(t, err)
	})

	t.Run("returns error for negative chunk size", func(t *testing.T) {
		_, err := SplitIntoChunks([]int{1, 2, 3}, -2)
		require.Error(t, err)
	})

	t.Run("returns empty result for nil slice", func(t *testing.T) {
		var slice []int
		slice = nil
		got, err := SplitIntoChunks(slice, 2)
		require.NoError(t, err)
		want := [][]int{}
		if cmp.Diff(got, want) != "" {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestBinarySearchRangeFunc(t *testing.T) {
	t.Run("returns the range of elements that are equal to the target", func(t *testing.T) {
		got := BinarySearchRangeFunc([]int{1, 2, 2, 2, 3, 4, 5, 6, 7, 8, 9}, 2, func(x, y int) int {
			return x - y
		})
		want := HalfOpenRange{Start: 1, End: 4}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	rapid.Check(t, func(t *rapid.T) {
		s := rapid.SliceOfN(rapid.IntRange(0, 10), 0, 10).Draw(t, "slice")
		slices.Sort(s)
		cmpFunc := stdcmp.Compare[int]
		target := rapid.IntRange(0, 10).Draw(t, "target")
		got := BinarySearchRangeFunc(s, target, cmpFunc)
		binarySearchFound := !got.IsEmpty()
		expected, linearSearchFound := linearSearchRangeFunc(s, target, cmpFunc)
		if binarySearchFound != linearSearchFound {
			t.Errorf("BinarySearchRangeFunc returned %v, but linearSearchRangeFunc returned %v", binarySearchFound, linearSearchFound)
		}
		if binarySearchFound && got != expected {
			t.Errorf("BinarySearchRangeFunc returned %v, but linearSearchRangeFunc returned %v", got, expected)
		}
	})
}

func linearSearchRangeFunc[S ~[]E, E, T any](x S, target T, cmp func(E, T) int) (HalfOpenRange, bool) {
	start := -1
	end := -1
	found := false
	for i, e := range x {
		if cmp(e, target) == 0 {
			if !found {
				found = true
				start = i
				end = i + 1
			} else {
				end = i + 1
			}
		}
	}
	if found {
		return HalfOpenRange{Start: start, End: end}, true
	}
	return HalfOpenRange{}, false
}
