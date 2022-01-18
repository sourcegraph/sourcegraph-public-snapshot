package result

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_searchRangeToHighlights(t *testing.T) {
	type testCase struct {
		input      string
		inputRange Range
		output     []HighlightedRange
	}

	cases := []testCase{
		{
			input: "abc",
			inputRange: Range{
				Start: Location{Offset: 1, Line: 0, Column: 1},
				End:   Location{Offset: 2, Line: 0, Column: 2},
			},
			output: []HighlightedRange{{
				Line:      0,
				Character: 1,
				Length:    1,
			}},
		},
		{
			input: "abc\ndef\n",
			inputRange: Range{
				Start: Location{Offset: 2, Line: 0, Column: 2},
				End:   Location{Offset: 5, Line: 1, Column: 1},
			},
			output: []HighlightedRange{{
				Line:      0,
				Character: 2,
				Length:    1,
			}, {
				Line:      1,
				Character: 0,
				Length:    1,
			}},
		},
		{
			input: "abc\ndef\nghi\n",
			inputRange: Range{
				Start: Location{Offset: 0, Line: 0, Column: 0},
				End:   Location{Offset: 11, Line: 2, Column: 3},
			},
			output: []HighlightedRange{{
				Line:      0,
				Character: 0,
				Length:    3,
			}, {
				Line:      1,
				Character: 0,
				Length:    3,
			}, {
				Line:      2,
				Character: 0,
				Length:    3,
			}},
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res := rangeToHighlights(tc.input, tc.inputRange)
			require.Equal(t, tc.output, res)
		})
	}
}
