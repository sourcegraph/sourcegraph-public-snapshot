package search

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type deltaResult struct {
	Delta Delta
	Hunks []hunkResult
}

type hunkResult struct {
	Hunk  Hunk
	Lines []Line
}

func TestDiffIter(t *testing.T) {
	diff := `a  b
@@ c
 d
-ef
+gh
 ij
@@ k
-lm
+no
p  q
@@ rs
 t
-u
+v
+w
`

	var results []deltaResult

	FormattedDiff([]byte(diff)).ForEachDelta(func(d Delta) bool {
		dr := deltaResult{Delta: d}
		d.ForEachHunk(func(h Hunk) bool {
			hunk := hunkResult{Hunk: h}
			h.ForEachLine(func(l Line) bool {
				hunk.Lines = append(hunk.Lines, l)
				return true
			})

			dr.Hunks = append(dr.Hunks, hunk)
			return true
		})

		results = append(results, dr)
		return true
	})

	expected := []deltaResult{{
		Delta: Delta{
			location: Location{},
			oldFile:  "a",
			newFile:  "b",
			hunks:    "@@ c\n d\n-ef\n+gh\n ij\n@@ k\n-lm\n+no\n",
		},
		Hunks: []hunkResult{{
			Hunk: Hunk{
				location: Location{Line: 1, Offset: 5},
				header:   "@@ c",
				lines:    " d\n-ef\n+gh\n ij\n",
			},
			Lines: []Line{
				{fullLine: " d\n", location: Location{Line: 2, Offset: 10}},
				{fullLine: "-ef\n", location: Location{Line: 3, Offset: 13}},
				{fullLine: "+gh\n", location: Location{Line: 4, Offset: 17}},
				{fullLine: " ij\n", location: Location{Line: 5, Offset: 21}},
			},
		}, {
			Hunk: Hunk{
				location: Location{Line: 6, Offset: 25},
				header:   "@@ k",
				lines:    "-lm\n+no\n",
			},
			Lines: []Line{
				{fullLine: "-lm\n", location: Location{Line: 7, Offset: 30}},
				{fullLine: "+no\n", location: Location{Line: 8, Offset: 34}},
			},
		}},
	}, {
		Delta: Delta{
			location: Location{Line: 9, Offset: 38},
			oldFile:  "p",
			newFile:  "q",
			hunks:    "@@ rs\n t\n-u\n+v\n+w\n",
		},
		Hunks: []hunkResult{{
			Hunk: Hunk{
				location: Location{Line: 10, Offset: 43},
				header:   "@@ rs",
				lines:    " t\n-u\n+v\n+w\n",
			},
			Lines: []Line{
				{fullLine: " t\n", location: Location{Line: 11, Offset: 49}},
				{fullLine: "-u\n", location: Location{Line: 12, Offset: 52}},
				{fullLine: "+v\n", location: Location{Line: 13, Offset: 55}},
				{fullLine: "+w\n", location: Location{Line: 14, Offset: 58}},
			},
		}},
	}}

	require.Equal(t, expected, results)
}
