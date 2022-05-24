package result

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMatches(t *testing.T) {
	// single line matches should always be roundtrippable
	t.Run("roundtrip", func(t *testing.T) {
		t.Run("multiline", func(t *testing.T) {
			cases := []MultilineMatch{{
				Preview: "abcd",
				Start:   LineColumn{0, 0},
				End:     LineColumn{0, 4},
			}, {
				Preview: "abcd",
				Start:   LineColumn{0, 0},
				End:     LineColumn{0, 3},
			}, {
				Preview: "abcd",
				Start:   LineColumn{3, 1},
				End:     LineColumn{3, 2},
			}}

			for _, tc := range cases {
				t.Run("", func(t *testing.T) {
					lineMatches := tc.AsLineMatches()
					require.Len(t, lineMatches, 1)
					multilineMatches := lineMatches[0].AsMultilineMatches()
					require.Len(t, multilineMatches, 1)
					require.Equal(t, tc, multilineMatches[0])
				})
			}
		})

		t.Run("oneline", func(t *testing.T) {
			cases := []*LineMatch{{
				Preview:          "abcd",
				LineNumber:       0,
				OffsetAndLengths: [][2]int32{{0, 4}},
			}, {
				Preview:          "abcd",
				LineNumber:       0,
				OffsetAndLengths: [][2]int32{{0, 3}},
			}, {
				Preview:          "abcd",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{1, 1}},
			}}

			for _, tc := range cases {
				t.Run("", func(t *testing.T) {
					multilineMatches := tc.AsMultilineMatches()
					require.Len(t, multilineMatches, 1)
					lineMatches := multilineMatches[0].AsLineMatches()
					require.Len(t, lineMatches, 1)
					require.Equal(t, tc, lineMatches[0])
				})
			}
		})
	})

	t.Run("AsLineMatches", func(t *testing.T) {
		cases := []struct {
			input  MultilineMatch
			output []*LineMatch
		}{{
			input: MultilineMatch{
				Preview: "line1\nline2\nline3",
				Start:   LineColumn{1, 1},
				End:     LineColumn{3, 1},
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
				Start:   LineColumn{1, 1},
				End:     LineColumn{1, 3},
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

	t.Run("AsMultilineMatches", func(t *testing.T) {
		cases := []struct {
			input  LineMatch
			output []MultilineMatch
		}{{
			input: LineMatch{
				Preview:          "0.2.4.6.8.10.13.16.19",
				LineNumber:       42,
				OffsetAndLengths: [][2]int32{{2, 2}, {8, 5}},
			},
			output: []MultilineMatch{{
				Preview: "0.2.4.6.8.10.13.16.19",
				Start:   LineColumn{42, 2},
				End:     LineColumn{42, 4},
			}, {
				Preview: "0.2.4.6.8.10.13.16.19",
				Start:   LineColumn{42, 8},
				End:     LineColumn{42, 13},
			}},
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				require.Equal(t, tc.output, tc.input.AsMultilineMatches())
			})
		}
	})
}
