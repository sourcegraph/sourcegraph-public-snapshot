package search

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
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
	diff := `a b
@@ c
 d
-ef
+gh
 ij
@@ k
-lm
+no
p q
@@ rs
 t
-u
+v
+w
`

	var results []deltaResult
	type Range = protocol.Range
	type Location = protocol.Location

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
			Range:   Range{Location{}, Location{Line: 9, Offset: 37}},
			oldFile: "a",
			newFile: "b",
			hunks:   "@@ c\n d\n-ef\n+gh\n ij\n@@ k\n-lm\n+no\n",
		},
		Hunks: []hunkResult{{
			Hunk: Hunk{
				Range:  Range{Location{Line: 1, Offset: 4}, Location{Line: 6, Offset: 24}},
				header: "@@ c",
				lines:  " d\n-ef\n+gh\n ij\n",
			},
			Lines: []Line{
				{fullLine: " d\n", Range: Range{Location{Line: 2, Offset: 9}, Location{Line: 3, Offset: 12}}},
				{fullLine: "-ef\n", Range: Range{Location{Line: 3, Offset: 12}, Location{Line: 4, Offset: 16}}},
				{fullLine: "+gh\n", Range: Range{Location{Line: 4, Offset: 16}, Location{Line: 5, Offset: 20}}},
				{fullLine: " ij\n", Range: Range{Location{Line: 5, Offset: 20}, Location{Line: 6, Offset: 24}}},
			},
		}, {
			Hunk: Hunk{
				Range:  Range{Location{Line: 6, Offset: 24}, Location{Line: 9, Offset: 37}},
				header: "@@ k",
				lines:  "-lm\n+no\n",
			},
			Lines: []Line{
				{fullLine: "-lm\n", Range: Range{Location{Line: 7, Offset: 29}, Location{Line: 8, Offset: 33}}},
				{fullLine: "+no\n", Range: Range{Location{Line: 8, Offset: 33}, Location{Line: 9, Offset: 37}}},
			},
		}},
	}, {
		Delta: Delta{
			Range:   Range{Location{Line: 9, Offset: 37}, Location{Line: 15, Offset: 59}},
			oldFile: "p",
			newFile: "q",
			hunks:   "@@ rs\n t\n-u\n+v\n+w\n",
		},
		Hunks: []hunkResult{{
			Hunk: Hunk{
				Range:  Range{Location{Line: 10, Offset: 41}, Location{Line: 15, Offset: 59}},
				header: "@@ rs",
				lines:  " t\n-u\n+v\n+w\n",
			},
			Lines: []Line{
				{fullLine: " t\n", Range: Range{Location{Line: 11, Offset: 47}, Location{Line: 12, Offset: 50}}},
				{fullLine: "-u\n", Range: Range{Location{Line: 12, Offset: 50}, Location{Line: 13, Offset: 53}}},
				{fullLine: "+v\n", Range: Range{Location{Line: 13, Offset: 53}, Location{Line: 14, Offset: 56}}},
				{fullLine: "+w\n", Range: Range{Location{Line: 14, Offset: 56}, Location{Line: 15, Offset: 59}}},
			},
		}},
	}}

	require.Equal(t, expected, results)
}
