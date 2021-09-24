package search

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func Test_matchesToRanges(t *testing.T) {
	type testCase struct {
		input          string
		matches        [][]int
		expectedRanges protocol.Ranges
	}

	cases := [...]testCase{
		0: {
			input:   "abc",
			matches: [][]int{{0, 1}, {2, 3}},
			expectedRanges: protocol.Ranges{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: protocol.Location{Offset: 2, Line: 0, Column: 2},
				End:   protocol.Location{Offset: 3, Line: 0, Column: 3},
			}},
		},
		1: {
			input:   "a\nc",
			matches: [][]int{{0, 1}, {2, 3}},
			expectedRanges: protocol.Ranges{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: protocol.Location{Offset: 2, Line: 1, Column: 0},
				End:   protocol.Location{Offset: 3, Line: 1, Column: 1},
			}},
		},
		2: {
			input:   "a\n\nc",
			matches: [][]int{{0, 1}, {3, 4}},
			expectedRanges: protocol.Ranges{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 1, Line: 0, Column: 1},
			}, {
				Start: protocol.Location{Offset: 3, Line: 2, Column: 0},
				End:   protocol.Location{Offset: 4, Line: 2, Column: 1},
			}},
		},
		3: {
			input:   "a\nb\nc\n",
			matches: [][]int{{0, 3}, {4, 5}},
			expectedRanges: protocol.Ranges{{
				Start: protocol.Location{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Location{Offset: 3, Line: 1, Column: 1},
			}, {
				Start: protocol.Location{Offset: 4, Line: 2, Column: 0},
				End:   protocol.Location{Offset: 5, Line: 2, Column: 1},
			}},
		},
		4: {
			input:   "abc\ndef\n",
			matches: [][]int{{1, 3}, {5, 7}},
			expectedRanges: protocol.Ranges{{
				Start: protocol.Location{Offset: 1, Line: 0, Column: 1},
				End:   protocol.Location{Offset: 3, Line: 0, Column: 3},
			}, {
				Start: protocol.Location{Offset: 5, Line: 1, Column: 1},
				End:   protocol.Location{Offset: 7, Line: 1, Column: 3},
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
