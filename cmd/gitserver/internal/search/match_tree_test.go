package search

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_matchesToRanges(t *testing.T) {
	type testCase struct {
		input          string
		matches        [][]int
		expectedRanges result.Ranges
	}

	cases := [...]testCase{
		0: {
			input:   "abc",
			matches: [][]int{{0, 1}, {2, 3}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 0, Column: 0},
				End:   result.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: result.Location{Offset: 2, Line: 0, Column: 2},
				End:   result.Location{Offset: 3, Line: 0, Column: 3},
			}},
		},
		1: {
			input:   "a\nc",
			matches: [][]int{{0, 1}, {2, 3}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 0, Column: 0},
				End:   result.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: result.Location{Offset: 2, Line: 1, Column: 0},
				End:   result.Location{Offset: 3, Line: 1, Column: 1},
			}},
		},
		2: {
			input:   "a\n\nc",
			matches: [][]int{{0, 1}, {3, 4}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 0, Column: 0},
				End:   result.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: result.Location{Offset: 3, Line: 2, Column: 0},
				End:   result.Location{Offset: 4, Line: 2, Column: 1},
			}},
		},
		3: {
			input:   "a\nb\nc\n",
			matches: [][]int{{0, 3}, {4, 5}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 0, Column: 0},
				End:   result.Location{Offset: 3, Line: 1, Column: 1},
			}, {
				Start: result.Location{Offset: 4, Line: 2, Column: 0},
				End:   result.Location{Offset: 5, Line: 2, Column: 1},
			}},
		},
		4: {
			input:   "abc\ndef\n",
			matches: [][]int{{1, 3}, {5, 7}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 1, Line: 0, Column: 1},
				End:   result.Location{Offset: 3, Line: 0, Column: 3},
			}, {
				Start: result.Location{Offset: 5, Line: 1, Column: 1},
				End:   result.Location{Offset: 7, Line: 1, Column: 3},
			}},
		},
		5: {
			input:   "›a", // three-byte unicode character
			matches: [][]int{{3, 4}},
			expectedRanges: result.Ranges{{
				Start: result.Location{Offset: 3, Line: 0, Column: 1},
				End:   result.Location{Offset: 4, Line: 0, Column: 2},
			}},
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ranges := matchesToRanges([]byte(tc.input), tc.matches)
			require.Equal(t, tc.expectedRanges, ranges)
		})
	}
}
