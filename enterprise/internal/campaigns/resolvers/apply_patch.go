package resolvers

import (
	"strings"

	"github.com/sourcegraph/go-diff/diff"
)

func applyPatch(fileContent string, fileDiff *diff.FileDiff) string {
	contentLines := strings.Split(fileContent, "\n")
	newContentLines := make([]string, 0)
	var lastLine int32 = 1
	// Assumes the hunks are sorted by ascending lines.
	for _, hunk := range fileDiff.Hunks {
		// Detect holes.
		if hunk.OrigStartLine != 0 && hunk.OrigStartLine != lastLine {
			originalLines := contentLines[lastLine-1 : hunk.OrigStartLine-1]
			newContentLines = append(newContentLines, originalLines...)
			lastLine += int32(len(originalLines))
		}
		hunkLines := strings.Split(string(hunk.Body), "\n")
		for _, line := range hunkLines {
			switch {
			case line == "":
				// Skip
			case strings.HasPrefix(line, "-"):
				lastLine++
			case strings.HasPrefix(line, "+"):
				newContentLines = append(newContentLines, line[1:])
			default:
				newContentLines = append(newContentLines, contentLines[lastLine-1])
				lastLine++
			}
		}
	}
	// Append remaining lines from original file.
	if origLines := int32(len(contentLines)); origLines > 0 && origLines != lastLine {
		newContentLines = append(newContentLines, contentLines[lastLine-1:]...)
	}
	return strings.Join(newContentLines, "\n")
}
