package codegraph

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/collections"
)

type IOccurrence interface {
	GetRange() []int32
}

func FindOccurrencesWithEqualRange[Occurrence IOccurrence](occurrences []Occurrence, search scip.Range) []Occurrence {
	interval := collections.BinarySearchRangeFunc(occurrences, search, func(occ Occurrence, r scip.Range) int {
		occRange := scip.NewRangeUnchecked(occ.GetRange())
		return occRange.CompareStrict(r)
	})
	return occurrences[interval.Start:interval.End]
}
