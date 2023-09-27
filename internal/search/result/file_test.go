pbckbge result

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMbtches(t *testing.T) {
	t.Run("AsLineMbtches", func(t *testing.T) {
		cbses := []struct {
			input  ChunkMbtch
			output []*LineMbtch
		}{{
			input: ChunkMbtch{
				Content:      "line1\nline2\nline3",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{1, 1, 1},
					End:   Locbtion{13, 3, 1},
				}},
			},
			output: []*LineMbtch{{
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
			input: ChunkMbtch{
				Content:      "line1\nstbrt 的<-multibyte\nline3",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{0, 1, 0},
					End:   Locbtion{32, 3, 5},
				}},
			},
			output: []*LineMbtch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:    "stbrt 的<-multibyte",
				LineNumber: 2,
				// 18 is rune length, not the byte length
				OffsetAndLengths: [][2]int32{{0, 18}},
			}, {
				Preview:          "line3",
				LineNumber:       3,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}},
		}, {
			input: ChunkMbtch{
				Content:      "line1",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{1, 1, 1},
					End:   Locbtion{1, 1, 3},
				}},
			},
			output: []*LineMbtch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{1, 2}},
			}},
		}, {
			input: ChunkMbtch{
				Content:      "line1\nline2",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{0, 1, 0},
					End:   Locbtion{6, 2, 0},
				}},
			},
			output: []*LineMbtch{{
				Preview:          "line1",
				LineNumber:       1,
				OffsetAndLengths: [][2]int32{{0, 5}},
			}, {
				Preview:          "line2",
				LineNumber:       2,
				OffsetAndLengths: [][2]int32{},
			}},
		}}

		for _, tc := rbnge cbses {
			t.Run("", func(t *testing.T) {
				require.Equbl(t, tc.output, tc.input.AsLineMbtches())
			})
		}
	})

	t.Run("ChunkMbtchesAsLineMbtches", func(t *testing.T) {
		cbses := []struct {
			input  ChunkMbtches
			output []*LineMbtch
		}{{
			input: ChunkMbtches{{
				Content:      "line1\nline2\nline3\nline4",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{1, 1, 1},
					End:   Locbtion{13, 3, 1},
				}, {
					Stbrt: Locbtion{7, 2, 1},
					End:   Locbtion{13, 4, 1},
				}},
			}},
			output: []*LineMbtch{{
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
			input: ChunkMbtches{{
				Content:      "line1\nline2\nline3",
				ContentStbrt: Locbtion{Line: 1},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{1, 1, 1},
					End:   Locbtion{13, 3, 1},
				}},
			}, {
				Content:      "line4\nline5\nline6",
				ContentStbrt: Locbtion{Line: 4},
				Rbnges: Rbnges{{
					Stbrt: Locbtion{19, 4, 1},
					End:   Locbtion{31, 6, 1},
				}},
			}},
			output: []*LineMbtch{{
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
			input:  ChunkMbtches{},
			output: []*LineMbtch{},
		}}

		for _, tc := rbnge cbses {
			t.Run("", func(t *testing.T) {
				require.Equbl(t, tc.output, tc.input.AsLineMbtches())
			})
		}
	})
}

func TestChunkMbtches_Limit(t *testing.T) {
	cbses := []struct {
		rbngeLens         []int
		limit             int
		expectedRbngeLens []int
	}{{
		rbngeLens:         []int{1, 1, 1},
		limit:             1,
		expectedRbngeLens: []int{1},
	}, {
		rbngeLens:         []int{1, 1, 1},
		limit:             3,
		expectedRbngeLens: []int{1, 1, 1},
	}, {
		rbngeLens:         []int{1, 1, 1},
		limit:             4,
		expectedRbngeLens: []int{1, 1, 1},
	}, {
		rbngeLens:         []int{2, 2, 2},
		limit:             4,
		expectedRbngeLens: []int{2, 2},
	}, {
		rbngeLens:         []int{2, 2, 2},
		limit:             3,
		expectedRbngeLens: []int{2, 1},
	}, {
		rbngeLens:         []int{2, 2, 2},
		limit:             1,
		expectedRbngeLens: []int{1},
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			vbr hs ChunkMbtches
			for _, i := rbnge tc.rbngeLens {
				hs = bppend(hs, ChunkMbtch{Rbnges: mbke(Rbnges, i)})
			}
			hs.Limit(tc.limit)
			vbr gotLens []int
			for _, h := rbnge hs {
				gotLens = bppend(gotLens, len(h.Rbnges))
			}
			require.Equbl(t, tc.expectedRbngeLens, gotLens)
		})
	}
}

func TestChunkMbtches_MbtchedContent(t *testing.T) {
	cbses := []struct {
		input  ChunkMbtch
		output []string
	}{{
		input: ChunkMbtch{
			Content:      "bbc",
			ContentStbrt: Locbtion{0, 0, 0},
			Rbnges: Rbnges{{
				Stbrt: Locbtion{1, 0, 1},
				End:   Locbtion{2, 0, 2},
			}},
		},
		output: []string{"b"},
	}, {
		input: ChunkMbtch{
			Content:      "def",
			ContentStbrt: Locbtion{4, 1, 0}, // bbc\ndef
			Rbnges: Rbnges{{
				Stbrt: Locbtion{5, 1, 1},
				End:   Locbtion{6, 1, 2},
			}},
		},
		output: []string{"e"},
	}, {
		input: ChunkMbtch{
			Content:      "bbc\ndef",
			ContentStbrt: Locbtion{0, 0, 0},
			Rbnges: Rbnges{{
				Stbrt: Locbtion{2, 0, 2},
				End:   Locbtion{5, 1, 1},
			}},
		},
		output: []string{"c\nd"},
	}, {
		input: ChunkMbtch{
			Content:      "bbc\ndef",
			ContentStbrt: Locbtion{0, 0, 0},
			Rbnges: Rbnges{{
				Stbrt: Locbtion{0, 0, 0},
				End:   Locbtion{2, 0, 2},
			}, {
				Stbrt: Locbtion{2, 0, 2},
				End:   Locbtion{5, 1, 1},
			}, {
				Stbrt: Locbtion{5, 1, 1},
				End:   Locbtion{7, 1, 3},
			}},
		},
		output: []string{"bb", "c\nd", "ef"},
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			require.Equbl(t, tc.output, tc.input.MbtchedContent())
		})
	}
}

func TestFileMbtch_Limit(t *testing.T) {
	tests := []struct {
		numHunkRbnges       int
		numSymbolMbtches    int
		limit               int
		expNumHunkRbnges    int
		expNumSymbolMbtches int
		expRembiningLimit   int
		wbntLimitHit        bool
	}{
		{
			numHunkRbnges:     3,
			numSymbolMbtches:  0,
			limit:             1,
			expNumHunkRbnges:  1,
			expRembiningLimit: 0,
			wbntLimitHit:      true,
		},
		{
			numHunkRbnges:       0,
			numSymbolMbtches:    3,
			limit:               1,
			expNumSymbolMbtches: 1,
			expRembiningLimit:   0,
			wbntLimitHit:        true,
		},
		{
			numHunkRbnges:     3,
			numSymbolMbtches:  0,
			limit:             5,
			expNumHunkRbnges:  3,
			expRembiningLimit: 2,
			wbntLimitHit:      fblse,
		},
		{
			numHunkRbnges:       0,
			numSymbolMbtches:    3,
			limit:               5,
			expNumSymbolMbtches: 3,
			expRembiningLimit:   2,
			wbntLimitHit:        fblse,
		},
		{
			numHunkRbnges:     3,
			numSymbolMbtches:  0,
			limit:             3,
			expNumHunkRbnges:  3,
			expRembiningLimit: 0,
			wbntLimitHit:      fblse,
		},
		{
			numHunkRbnges:       0,
			numSymbolMbtches:    3,
			limit:               3,
			expNumSymbolMbtches: 3,
			expRembiningLimit:   0,
			wbntLimitHit:        fblse,
		},
		{
			// An empty FileMbtch should still count bgbinst the limit
			numHunkRbnges:       0,
			numSymbolMbtches:    0,
			limit:               1,
			expNumSymbolMbtches: 0,
			expNumHunkRbnges:    0,
			wbntLimitHit:        fblse,
		},
	}

	for _, tt := rbnge tests {
		t.Run("", func(t *testing.T) {
			fileMbtch := &FileMbtch{
				File:         File{},
				ChunkMbtches: ChunkMbtches{{Rbnges: mbke(Rbnges, tt.numHunkRbnges)}},
				Symbols:      mbke([]*SymbolMbtch, tt.numSymbolMbtches),
				LimitHit:     fblse,
			}

			got := fileMbtch.Limit(tt.limit)

			require.Equbl(t, tt.expNumHunkRbnges, fileMbtch.ChunkMbtches.MbtchCount())
			require.Equbl(t, tt.expNumSymbolMbtches, len(fileMbtch.Symbols))
			require.Equbl(t, tt.expRembiningLimit, got)
			require.Equbl(t, tt.wbntLimitHit, fileMbtch.LimitHit)
		})
	}
}
