package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestToResultResolverList(t *testing.T) {
	content := "ab"

	git.Mocks.ReadFile = func(_ api.CommitID, _ string) ([]byte, error) {
		return []byte(content), nil
	}

	test := func(input string) string {
		computeQuery, _ := compute.Parse(input)
		resolvers, _ := toResultResolverList(
			context.Background(),
			computeQuery.Command,
			[]result.Match{&result.FileMatch{}},
			database.NewMockDB(),
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

	autogold.Want("resolver copies all match results", `["a","b"]`).Equal(t, test("a|b"))
}
