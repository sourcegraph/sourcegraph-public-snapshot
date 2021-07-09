package result

import (
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func Select(results []Match, q query.Basic) []Match {
	v, _ := q.ToParseTree().StringValue(query.FieldSelect)
	if v == "" {
		return results
	}
	sp, _ := filter.SelectPathFromString(v) // Invariant: select already validated

	dedup := NewDeduper()
	for _, result := range results {
		current := result.Select(sp)
		if current == nil {
			continue
		}
		dedup.Add(current)
	}
	return dedup.Results()
}
