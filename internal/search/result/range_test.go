pbckbge result

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sebrchRbngeToHighlights(t *testing.T) {
	type testCbse struct {
		input      string
		inputRbnge Rbnge
		output     []HighlightedRbnge
	}

	cbses := []testCbse{
		{
			input: "bbc",
			inputRbnge: Rbnge{
				Stbrt: Locbtion{Offset: 1, Line: 0, Column: 1},
				End:   Locbtion{Offset: 2, Line: 0, Column: 2},
			},
			output: []HighlightedRbnge{{
				Line:      0,
				Chbrbcter: 1,
				Length:    1,
			}},
		},
		{
			input: "bbc\ndef\n",
			inputRbnge: Rbnge{
				Stbrt: Locbtion{Offset: 2, Line: 0, Column: 2},
				End:   Locbtion{Offset: 5, Line: 1, Column: 1},
			},
			output: []HighlightedRbnge{{
				Line:      0,
				Chbrbcter: 2,
				Length:    1,
			}, {
				Line:      1,
				Chbrbcter: 0,
				Length:    1,
			}},
		},
		{
			input: "bbc\ndef\nghi\n",
			inputRbnge: Rbnge{
				Stbrt: Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   Locbtion{Offset: 11, Line: 2, Column: 3},
			},
			output: []HighlightedRbnge{{
				Line:      0,
				Chbrbcter: 0,
				Length:    3,
			}, {
				Line:      1,
				Chbrbcter: 0,
				Length:    3,
			}, {
				Line:      2,
				Chbrbcter: 0,
				Length:    3,
			}},
		},
	}

	for i, tc := rbnge cbses {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			res := rbngeToHighlights(tc.input, tc.inputRbnge)
			require.Equbl(t, tc.output, res)
		})
	}
}
