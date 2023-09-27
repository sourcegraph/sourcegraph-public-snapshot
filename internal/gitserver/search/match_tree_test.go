pbckbge sebrch

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func Test_mbtchesToRbnges(t *testing.T) {
	type testCbse struct {
		input          string
		mbtches        [][]int
		expectedRbnges result.Rbnges
	}

	cbses := [...]testCbse{
		0: {
			input:   "bbc",
			mbtches: [][]int{{0, 1}, {2, 3}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   result.Locbtion{Offset: 1, Line: 0, Column: 1},
			}, {
				Stbrt: result.Locbtion{Offset: 2, Line: 0, Column: 2},
				End:   result.Locbtion{Offset: 3, Line: 0, Column: 3},
			}},
		},
		1: {
			input:   "b\nc",
			mbtches: [][]int{{0, 1}, {2, 3}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   result.Locbtion{Offset: 1, Line: 0, Column: 1},
			}, {
				Stbrt: result.Locbtion{Offset: 2, Line: 1, Column: 0},
				End:   result.Locbtion{Offset: 3, Line: 1, Column: 1},
			}},
		},
		2: {
			input:   "b\n\nc",
			mbtches: [][]int{{0, 1}, {3, 4}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   result.Locbtion{Offset: 1, Line: 0, Column: 1},
			}, {
				Stbrt: result.Locbtion{Offset: 3, Line: 2, Column: 0},
				End:   result.Locbtion{Offset: 4, Line: 2, Column: 1},
			}},
		},
		3: {
			input:   "b\nb\nc\n",
			mbtches: [][]int{{0, 3}, {4, 5}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   result.Locbtion{Offset: 3, Line: 1, Column: 1},
			}, {
				Stbrt: result.Locbtion{Offset: 4, Line: 2, Column: 0},
				End:   result.Locbtion{Offset: 5, Line: 2, Column: 1},
			}},
		},
		4: {
			input:   "bbc\ndef\n",
			mbtches: [][]int{{1, 3}, {5, 7}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 1, Line: 0, Column: 1},
				End:   result.Locbtion{Offset: 3, Line: 0, Column: 3},
			}, {
				Stbrt: result.Locbtion{Offset: 5, Line: 1, Column: 1},
				End:   result.Locbtion{Offset: 7, Line: 1, Column: 3},
			}},
		},
		5: {
			input:   "›b", // three-byte unicode chbrbcter
			mbtches: [][]int{{3, 4}},
			expectedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 3, Line: 0, Column: 1},
				End:   result.Locbtion{Offset: 4, Line: 0, Column: 2},
			}},
		},
	}

	for i, tc := rbnge cbses {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			rbnges := mbtchesToRbnges([]byte(tc.input), tc.mbtches)
			require.Equbl(t, tc.expectedRbnges, rbnges)
		})
	}
}
