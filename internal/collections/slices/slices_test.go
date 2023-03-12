package slices

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Min(t *testing.T) {
	t.Run("Returns first int that is smaller", func(t *testing.T) {
		got := Min(1, 2)
		want := 1
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Returns second int that is smaller", func(t *testing.T) {
		got := Min(2, 1)
		want := 1
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Works with a float as well", func(t *testing.T) {
		got := Min(1.5, 1.52)
		want := 1.5
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Works with infinity", func(t *testing.T) {
		got := Min(1.5, math.Inf(1))
		want := 1.5
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Works with negative infinity", func(t *testing.T) {
		got := Min(1.5, math.Inf(-1))
		want := math.Inf(-1)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func Test_Chunk(t *testing.T) {
	tests := map[string]struct {
		inputSlice []any
		chunkSize  int
		want       [][]any
	}{
		"splits a slice into chunks of size 3": {
			inputSlice: []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			chunkSize:  3,
			want:       [][]any{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		"type of slice does not matter": {
			inputSlice: []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			chunkSize:  4,
			want:       [][]any{{"a", "b", "c", "d"}, {"e", "f", "g", "h"}, {"i", "j"}},
		},
		"splits into 1 chunk if slice is smaller than requested chunk size": {
			inputSlice: []any{1, 2, 3},
			chunkSize:  4,
			want:       [][]any{{1, 2, 3}},
		},
		"works with chunk size of 1": {
			inputSlice: []any{1, 2, 3},
			chunkSize:  1,
			want:       [][]any{{1}, {2}, {3}},
		},
		"works with empty slice": {
			inputSlice: []any{},
			chunkSize:  4,
			want:       [][]any{{}},
		},
		"works with nil slice": {
			inputSlice: nil,
			chunkSize:  2,
			want:       [][]any{nil},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(test.want, Chunk(test.inputSlice, test.chunkSize)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
