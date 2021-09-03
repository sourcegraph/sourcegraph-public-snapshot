package search

import (
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	stream "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// segmentToRangs converts line match ranges into absolute ranges.
func segmentToRanges(lineNumber int, segments [][2]int32) []stream.Range {
	ranges := make([]stream.Range, 0, len(segments))
	for _, segment := range segments {
		ranges = append(ranges, stream.Range{
			Start: stream.Location{
				Offset: -1,
				Line:   lineNumber,
				Column: int(segment[0]),
			},
			End: stream.Location{
				Offset: -1,
				Line:   lineNumber,
				Column: int(segment[0]) + int(segment[1]),
			},
		})
	}
	return ranges
}

// group is a list of contiguous line matches by line number.
type group []*result.LineMatch

// toHunk converts a group of line matches to a hunk. A hunk comprises:
// (1) file `Content` (decorated for rendering depending on the request) that
// spans `LineStart + LineCount`, and
// (2) a list `Matches` which specify ranges of values to emphasize specially
// (e.g., with overlay-highlights) within the hunk range.
func toHunk(group group) stream.DecoratedHunk {
	matches := make([]stream.Range, 0, len(group))
	for _, line := range group {
		matches = append(matches, segmentToRanges(int(line.LineNumber), line.OffsetAndLengths)...)
	}
	return stream.DecoratedHunk{
		Content:   stream.DecoratedContent{Plaintext: "Placeholder"}, // TODO(rvantonder): populate this with decorated content
		LineStart: int(group[0].LineNumber),
		LineCount: len(group),
		Matches:   matches,
	}
}

// groupLineMatches converts a flat slice of line matches to groups of
// contiguous line matches based on line numbers.
func groupLineMatches(lineMatches []*result.LineMatch) []group {
	var groups []group
	var previousLine *result.LineMatch
	var currentGroup group
	for _, line := range lineMatches {
		if previousLine == nil {
			previousLine = line
		}
		if len(currentGroup) == 0 {
			currentGroup = append(currentGroup, line)
			// Invariant: previousLine is set to first line match.
			continue
		}
		if line.LineNumber-1 == previousLine.LineNumber {
			currentGroup = append(currentGroup, line)
			previousLine = line
			continue
		}
		groups = append(groups, currentGroup)
		currentGroup = group{line}
		previousLine = line
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}
