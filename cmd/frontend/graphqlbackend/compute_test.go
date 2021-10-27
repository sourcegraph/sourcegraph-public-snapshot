package graphqlbackend

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestToResultResolverList(t *testing.T) {
	matches := []result.Match{
		&result.FileMatch{
			LineMatches: []*result.LineMatch{
				{Preview: "a"},
				{Preview: "b"},
			},
		},
	}
	test := func(input string) string {
		computeQuery, _ := compute.Parse(input)
		resolvers, _ := toResultResolverList(
			context.Background(),
			computeQuery.Command,
			matches,
			new(dbtesting.MockDB),
		)
		var results []string
		for _, r := range resolvers {
			for _, m := range r.result.(*computeMatchContextResolver).matches {
				results = append(results, m.Value())
			}
		}
		v, _ := json.Marshal(results)
		return string(v)
	}

	autogold.Want("resolver copies all match results", `["a","b"]`).Equal(t, test("a|b"))
}
