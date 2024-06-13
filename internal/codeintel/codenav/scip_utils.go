package codenav

import (
	"slices"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

type IOccurrence interface {
	GetRange() []int32
}

func findOccurrencesWithEqualRange[Occurrence IOccurrence](occurrences []Occurrence, search scip.Range) []Occurrence {
	n, found := slices.BinarySearchFunc(occurrences, search.Start, func(occ Occurrence, p scip.Position) int {
		occRange := scip.NewRangeUnchecked(occ.GetRange())
		return occRange.Start.Compare(p)
	})
	results := []Occurrence{}
	if !found {
		return results
	}
	// Binary search is not guaranteed to find the last or first index, so we need to check in both directions
	for _, occurrence := range occurrences[n:] {
		parsedRange := scip.NewRangeUnchecked(occurrence.GetRange())
		if parsedRange.Compare(search) == 0 {
			results = append(results, occurrence)
		} else {
			break
		}
	}
	for i := n - 1; i >= 0; i-- {
		occurrence := occurrences[i]
		parsedRange := scip.NewRangeUnchecked(occurrence.GetRange())
		if parsedRange.Compare(search) == 0 {
			results = append(results, occurrence)
		} else {
			break
		}
	}

	return results
}
