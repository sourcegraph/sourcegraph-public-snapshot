package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestToResultResolverList(t *testing.T) {
	test := func(input string, matches []result.Match) string {
		computeQuery, _ := compute.Parse(input)
		resolvers, _ := toResultResolverList(
			context.Background(),
			computeQuery.Command,
			matches,
			dbmocks.NewMockDB(),
		)
		results := make([]string, 0, len(resolvers))
		for _, r := range resolvers {
			if rr, ok := r.ToComputeMatchContext(); ok {
				matches := rr.Matches()
				for _, m := range matches {
					results = append(results, m.Value())
				}
			}
		}
		v, _ := json.Marshal(results)
		return string(v)
	}

	nonNilMatches := []result.Match{
		&result.FileMatch{
			ChunkMatches: result.ChunkMatches{{
				Content: "a",
				Ranges: result.Ranges{{
					Start: result.Location{Offset: 0, Line: 1, Column: 0},
					End:   result.Location{Offset: 1, Line: 1, Column: 1},
				}},
			}, {
				Content: "b",
				Ranges: result.Ranges{{
					Start: result.Location{Offset: 0, Line: 2, Column: 0},
					End:   result.Location{Offset: 1, Line: 2, Column: 1},
				}},
			}},
		},
	}
	autogold.Expect(`["a","b"]`).Equal(t, test("a|b", nonNilMatches))

	producesNilResult := []result.Match{&result.CommitMatch{}}
	autogold.Expect("[]").Equal(t, test("a|b", producesNilResult))
}
