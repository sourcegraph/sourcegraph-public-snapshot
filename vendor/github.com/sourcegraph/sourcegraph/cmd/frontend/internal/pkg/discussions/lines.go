package discussions

import (
	"strings"
)

// LineRange represents a line range in a file.
type LineRange struct {
	// StarLine of the range (zero-based, inclusive).
	StartLine int

	// EndLine of the range (zero-based, exclusive).
	EndLine int
}

// LinesForSelection returns the lines from the given file's contents for the
// given selection.
func LinesForSelection(fileContent string, selection LineRange) (linesBefore, lines, linesAfter []string) {
	allLines := strings.Split(fileContent, "\n")
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		} else if v > max {
			return max
		}
		return v
	}
	linesForRange := func(startLine, endLine int) []string {
		startLine = clamp(startLine, 0, len(allLines))
		endLine = clamp(endLine, 0, len(allLines))
		selectedLines := allLines[startLine:endLine]
		return selectedLines
	}
	linesBefore = linesForRange(selection.StartLine-3, selection.StartLine)
	lines = linesForRange(selection.StartLine, selection.EndLine)
	linesAfter = linesForRange(selection.EndLine, selection.EndLine+3)
	return
}
