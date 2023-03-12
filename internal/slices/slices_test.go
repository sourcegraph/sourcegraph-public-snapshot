package slices

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMin(t *testing.T) {
	t.Run("TestInts", func(t *testing.T) {
		tests := map[string]struct {
			input []int
			want  int
		}{
			"returns first int that is smaller": {
				input: []int{1, 2},
				want:  1,
			},
			"returns second int that is smaller": {
				input: []int{2, -1},
				want:  -1,
			},
			"if only 1 int, returns it": {
				input: []int{5},
				want:  5,
			},
			"longer list": {
				input: []int{1, 5, -6, 10, -99, 100},
				want:  -99,
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				var got int
				if len(test.input) == 1 {
					got = Min(test.input[0])
				} else {
					got = Min(test.input[0], test.input[1:]...)
				}
				if got != test.want {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			})
		}
	})

	t.Run("TestFloats", func(t *testing.T) {
		tests := map[string]struct {
			input []float64
			want  float64
		}{
			"returns smaller float": {
				input: []float64{1.5, 1.52},
				want:  1.5,
			},
			"works with infinity": {
				input: []float64{1.5, math.Inf(1)},
				want:  1.5,
			},
			"works with negative infinity": {
				input: []float64{1.5, math.Inf(-1)},
				want:  math.Inf(-1),
			},
			"single input returns input": {
				input: []float64{0.234},
				want:  0.234,
			},
			"longer list works": {
				input: []float64{0.2, -9.5, 100, -5.3},
				want:  -9.5,
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				var got float64
				if len(test.input) == 1 {
					got = Min(test.input[0])
				} else {
					got = Min(test.input[0], test.input[1:]...)
				}
				if got != test.want {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			})
		}
	})
}

func TestChunk(t *testing.T) {
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
