package commit

import (
	"strconv"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestCommitSearchResult_Limit(t *testing.T) {
	f := func(nHighlights []int, limitInput uint32) bool {
		cr := &result.CommitMatch{
			Body: result.HighlightedString{
				Highlights: make([]result.HighlightedRange, len(nHighlights)),
			},
		}

		// It isn't interesting to test limit > ResultCount, so we bound it to
		// [1, ResultCount]
		count := cr.ResultCount()
		limit := (int(limitInput) % count) + 1

		after := cr.Limit(limit)
		newCount := cr.ResultCount()

		if after == 0 && newCount == limit {
			return true
		}

		t.Logf("failed limit=%d count=%d => after=%d newCount=%d", limit, count, after, newCount)
		return false
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error("quick check failed")
	}

	for nSymbols := 0; nSymbols <= 3; nSymbols++ {
		for limit := 0; limit <= nSymbols; limit++ {
			if !f(make([]int, nSymbols), uint32(limit)) {
				t.Error("small exhaustive check failed")
			}
		}
	}
}

func Test_searchRangeToHighlights(t *testing.T) {
	type testCase struct {
		input      string
		inputRange result.Range
		output     []result.HighlightedRange
	}

	cases := []testCase{
		{
			input: "abc",
			inputRange: result.Range{
				Start: result.Location{Offset: 1, Line: 0, Column: 1},
				End:   result.Location{Offset: 2, Line: 0, Column: 2},
			},
			output: []result.HighlightedRange{{
				Line:      0,
				Character: 1,
				Length:    1,
			}},
		},
		{
			input: "abc\ndef\n",
			inputRange: result.Range{
				Start: result.Location{Offset: 2, Line: 0, Column: 2},
				End:   result.Location{Offset: 5, Line: 1, Column: 1},
			},
			output: []result.HighlightedRange{{
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
			inputRange: result.Range{
				Start: result.Location{Offset: 0, Line: 0, Column: 0},
				End:   result.Location{Offset: 11, Line: 2, Column: 3},
			},
			output: []result.HighlightedRange{{
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
			res := searchRangeToHighlights(tc.input, tc.inputRange)
			require.Equal(t, tc.output, res)
		})
	}
}
