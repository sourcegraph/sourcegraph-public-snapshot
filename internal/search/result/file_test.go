package result

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMatches(t *testing.T) {
	t.Run("AsLineMatches", func(t *testing.T) {
		cases := []struct {
			input  HunkMatch
			output []*LineMatch
		}{{
			input: HunkMatch{
				Content:      "line1\nline2\nline3",
				ContentStart: Location{Line: 1},
				Ranges: Ranges{{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				}},
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
			input: HunkMatch{
				Content:      "line1\nstart 的<-multibyte\nline3",
				ContentStart: Location{Line: 1},
				Ranges: Ranges{{
					Start: Location{0, 1, 0},
					End:   Location{32, 3, 5},
				}},
			},
			output: []*LineMatch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:    "start 的<-multibyte",
				LineNumber: 2,
				// 18 is rune length, not the byte length
				OffsetAndLengths: [][2]int32{{0, 18}},
			}, {
				Preview:          "line3",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}},
		}, {
			input: HunkMatch{
				Content:      "line1",
				ContentStart: Location{Line: 1},
				Ranges: Ranges{{
					Start: Location{1, 1, 1},
					End:   Location{1, 1, 3},
				}},
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
				require.Equal(t, tc.output, tc.input.AsLineMatches())
			})
		}
	})

	t.Run("HunkMatchesAsLineMatches", func(t *testing.T) {
		cases := []struct {
			input  HunkMatches
			output []*LineMatch
		}{{
			input: HunkMatches{{
				Content:      "line1\nline2\nline3\nline4",
				ContentStart: Location{Line: 1},
				Ranges: Ranges{{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				}, {
					Start: Location{7, 2, 1},
					End:   Location{13, 4, 1},
				}},
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
			input: HunkMatches{{
				Content:      "line1\nline2\nline3",
				ContentStart: Location{Line: 1},
				Ranges: Ranges{{
					Start: Location{1, 1, 1},
					End:   Location{13, 3, 1},
				}},
			}, {
				Content:      "line4\nline5\nline6",
				ContentStart: Location{Line: 4},
				Ranges: Ranges{{
					Start: Location{19, 4, 1},
					End:   Location{31, 6, 1},
				}},
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
			input:  HunkMatches{},
			output: []*LineMatch{},
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				require.Equal(t, tc.output, tc.input.AsLineMatches())
			})
		}
	})
}

func TestHunkMatches_Limit(t *testing.T) {
	cases := []struct {
		rangeLens         []int
		limit             int
		expectedRangeLens []int
	}{{
		rangeLens:         []int{1, 1, 1},
		limit:             1,
		expectedRangeLens: []int{1},
	}, {
		rangeLens:         []int{1, 1, 1},
		limit:             3,
		expectedRangeLens: []int{1, 1, 1},
	}, {
		rangeLens:         []int{1, 1, 1},
		limit:             4,
		expectedRangeLens: []int{1, 1, 1},
	}, {
		rangeLens:         []int{2, 2, 2},
		limit:             4,
		expectedRangeLens: []int{2, 2},
	}, {
		rangeLens:         []int{2, 2, 2},
		limit:             3,
		expectedRangeLens: []int{2, 1},
	}, {
		rangeLens:         []int{2, 2, 2},
		limit:             1,
		expectedRangeLens: []int{1},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			var hs HunkMatches
			for _, i := range tc.rangeLens {
				hs = append(hs, HunkMatch{Ranges: make(Ranges, i)})
			}
			hs.Limit(tc.limit)
			var gotLens []int
			for _, h := range hs {
				gotLens = append(gotLens, len(h.Ranges))
			}
			require.Equal(t, tc.expectedRangeLens, gotLens)
		})
	}
}

func TestHunkMatches_MatchedContent(t *testing.T) {
	cases := []struct {
		input  HunkMatch
		output []string
	}{{
		input: HunkMatch{
			Content:      "abc",
			ContentStart: Location{0, 0, 0},
			Ranges: Ranges{{
				Start: Location{1, 0, 1},
				End:   Location{2, 0, 2},
			}},
		},
		output: []string{"b"},
	}, {
		input: HunkMatch{
			Content:      "def",
			ContentStart: Location{4, 1, 0}, // abc\ndef
			Ranges: Ranges{{
				Start: Location{5, 1, 1},
				End:   Location{6, 1, 2},
			}},
		},
		output: []string{"e"},
	}, {
		input: HunkMatch{
			Content:      "abc\ndef",
			ContentStart: Location{0, 0, 0},
			Ranges: Ranges{{
				Start: Location{2, 0, 2},
				End:   Location{5, 1, 1},
			}},
		},
		output: []string{"c\nd"},
	}, {
		input: HunkMatch{
			Content:      "abc\ndef",
			ContentStart: Location{0, 0, 0},
			Ranges: Ranges{{
				Start: Location{0, 0, 0},
				End:   Location{2, 0, 2},
			}, {
				Start: Location{2, 0, 2},
				End:   Location{5, 1, 1},
			}, {
				Start: Location{5, 1, 1},
				End:   Location{7, 1, 3},
			}},
		},
		output: []string{"ab", "c\nd", "ef"},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.output, tc.input.MatchedContent())
		})
	}
}
