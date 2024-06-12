package codenav

import (
	"slices"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

type IOccurrence interface {
	GetRange() []int32
}

func findIntersectingOccurrences[Occurrence IOccurrence](occurrences []Occurrence, search scip.Range) []Occurrence {
	n, _ := slices.BinarySearchFunc(occurrences, search.Start, func(occ Occurrence, p scip.Position) int {
		occRange := scip.NewRangeUnchecked(occ.GetRange())
		return occRange.Start.Compare(p)
	})
	n = max(0, n-1)

	result := make([]Occurrence, 0)
	for _, occurrence := range occurrences[n:] {
		parsedRange := scip.NewRangeUnchecked(occurrence.GetRange())
		if search.End.Compare(parsedRange.Start) < 0 {
			break
		}
		if search.Intersects(parsedRange) {
			result = append(result, occurrence)
		}
	}
	return result
}
