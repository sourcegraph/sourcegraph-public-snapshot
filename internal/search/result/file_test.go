package result

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMatches(t *testing.T) {
	t.Run("AsLineMatches", func(t *testing.T) {
		cases := []struct {
			input  MultilineMatch
			output []*LineMatch
		}{{
			input: MultilineMatch{
				Preview: "line1\nline2\nline3",
				Range: Range{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				},
			},
			output: []*LineMatch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{1, 4}},
			}, {
				Preview:          "line2",
				LineNumber:       2,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:          "line3",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{0, 1}},
			}},
		}, {
			input: MultilineMatch{
				Preview: "line1",
				Range: Range{
					Start: Location{1, 1, 1},
					End:   Location{1, 1, 3},
				},
			},
			output: []*LineMatch{
				{
					Preview:          "line1",
					LineNumber:       1,
					OffsetAndLengths: [][2]int32{{1, 2}},
				},
			},
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				require.Equal(t, tc.input.AsLineMatches(), tc.output)
			})
		}
	})

	t.Run("MultilineSliceAsLineMatchSlice", func(t *testing.T) {
		cases := []struct {
			input  []MultilineMatch
			output []*LineMatch
		}{{
			input: []MultilineMatch{{
				Preview: "line1\nline2\nline3",
				Range: Range{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				},
			}, {
				Preview: "line2\nline3\nline4",
				Range: Range{
					Start: Location{7, 2, 1},
					End:   Location{13, 4, 1},
				},
			}},
			output: []*LineMatch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{1, 4}},
			}, {
				Preview:          "line2",
				LineNumber:       2,
				OffsetAndLengths: [][2]int32{{0, 5}, {1, 4}},
			}, {
				Preview:          "line3",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{0, 1}, {0, 5}},
			}, {
				Preview:          "line4",
				LineNumber:       4,
				OffsetAndLengths: [][2]int32{{0, 1}},
			}},
		}, {
			input: []MultilineMatch{{
				Preview: "line1\nline2\nline3",
				Range: Range{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				},
			}, {
				Preview: "line4\nline5\nline6",
				Range: Range{
					Start: Location{19, 4, 1},
					End:   Location{31, 6, 1},
				},
			}},
			output: []*LineMatch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{1, 4}},
			}, {
				Preview:          "line2",
				LineNumber:       2,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:          "line3",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{0, 1}},
			}, {
				Preview:          "line4",
				LineNumber:       4,
				OffsetAndLengths: [][2]int32{{1, 4}},
			}, {
				Preview:          "line5",
				LineNumber:       5,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:          "line6",
				LineNumber:       6,
				OffsetAndLengths: [][2]int32{{0, 1}},
			}},
		}, {
			input:  []MultilineMatch{},
			output: []*LineMatch{},
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				require.Equal(t, MultilineSliceAsLineMatchSlice(tc.input), tc.output)
			})
		}
	})
}
