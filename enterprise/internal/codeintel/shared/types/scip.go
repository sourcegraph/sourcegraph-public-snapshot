package types

import (
	"sort"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// FindOccurrences filters the given occurrences and returns those that contain the position
// constructed from line and character. The order of the output slice is "outside-in", so that
// earlier occurrences properly enclose later occurrences.
func FindOccurrences(occurrences []*scip.Occurrence, line, character int) []*scip.Occurrence {
	var filtered []*scip.Occurrence
	for _, o := range occurrences {
		if comparePositionSCIP(scip.NewRange(o.Range), line, character) == 0 {
			filtered = append(filtered, o)
		}
	}

	return SortOccurrences(filtered)
}

// TODO - check ordering condition
func SortOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	sort.Slice(occurrences, func(i, j int) bool {
		ri := scip.NewRange(occurrences[i].Range)
		rj := scip.NewRange(occurrences[j].Range)

		return comparePositionSCIP(ri, int(rj.Start.Line), int(rj.Start.Character)) != 0
	})

	return occurrences
}

// TODO - check ordering condition
func SortRanges(ranges []*scip.Range) []*scip.Range {
	sort.Slice(ranges, func(i, j int) bool {
		return comparePositionSCIP(ranges[i], int(ranges[j].Start.Line), int(ranges[j].Start.Character)) != 0
	})

	return ranges
}

// comparePositionSCIP compares the range r with the position constructed from line and character.
// Returns -1 if the position occurs before the range, +1 if it occurs after, and 0 if the
// position is inside of the range.
func comparePositionSCIP(r *scip.Range, line, character int) int {
	if line < int(r.Start.Line) {
		return 1
	}

	if line > int(r.End.Line) {
		return -1
	}

	if line == int(r.Start.Line) && character < int(r.Start.Character) {
		return 1
	}

	if line == int(r.End.Line) && character >= int(r.End.Character) {
		return -1
	}

	return 0
}
