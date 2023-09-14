package priority

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
)

const (
	Simple        = LiteralCost
	Slow          = RegexpCost
	Long          = StructuralCost
	LikelyTimeout = StructuralCost * 10
) // values that could associate a speed to a floating point

func TestQueryAnalyzerCost(t *testing.T) {
	defaultHandlers := []CostHeuristic{QueryCost, RepositoriesCost}

	testCases := []struct {
		name                   string
		query1                 string
		numberOfRepositoriesQ1 int64
		repositoryByteSizesQ1  []int64
		query2                 string
		numberOfRepositoriesQ2 int64
		repositoryByteSizesQ2  []int64
		compare                assert.ComparisonAssertionFunc
		handlers               []CostHeuristic
	}{
		{
			name:     "literal diff query should be more than literal query ",
			query1:   "insights",
			query2:   "Type:diff insights",
			compare:  assert.Less,
			handlers: defaultHandlers,
		},
		{
			name:     "literal diff query with author should reduce complexity",
			query1:   "type:diff author:someone insights",
			query2:   "type:diff insights",
			compare:  assert.Less,
			handlers: defaultHandlers,
		},
		{
			name:     "a filter should reduce complexity",
			query1:   "patterntype:regexp [0-9]+ lang:go",
			query2:   "patterntype:regexp [0-9]+",
			compare:  assert.Less,
			handlers: defaultHandlers,
		},
		{
			name:     "multiple filters further reduces complexity",
			query1:   "file:insights lang:go DashboardResolver",
			query2:   "lang:go DashboardResolver",
			compare:  assert.Less,
			handlers: defaultHandlers,
		},
		{
			name:                   "small difference in num repos no difference",
			query1:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ1: 1,
			query2:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ2: 5,
			handlers:               defaultHandlers,
			compare:                assert.Equal,
		},
		{
			name:                   "large difference in num repos makes difference",
			query1:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ1: 1,
			query2:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ2: 20000,
			handlers:               defaultHandlers,
			compare:                assert.Less,
		},
		{
			name:                   "num repos continues to scale",
			query1:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ1: 20000,
			query2:                 "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositoriesQ2: 40000,
			handlers:               defaultHandlers,
			compare:                assert.Less,
		},
		{
			name:                   "queries over larege repos add complexity",
			query1:                 "patterntype:structural [a] archive:yes fork:yes index:no",
			numberOfRepositoriesQ1: 3,
			repositoryByteSizesQ1:  []int64{100, 100, 100},
			query2:                 "patterntype:structural [a] archive:yes fork:yes index:no",
			numberOfRepositoriesQ2: 3,
			repositoryByteSizesQ2:  []int64{100, megarepoSizeThreshold, gigarepoSizeThreshold},
			handlers:               defaultHandlers,
			compare:                assert.Less,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryAnalyzer := NewQueryAnalyzer(tc.handlers...)
			queryPlan1, err := querybuilder.ParseQuery(tc.query1, "literal")
			if err != nil {
				t.Fatal(err)
			}
			queryPlan2, err := querybuilder.ParseQuery(tc.query2, "literal")
			if err != nil {
				t.Fatal(err)
			}
			cost1 := queryAnalyzer.Cost(&QueryObject{
				Query:                queryPlan1,
				NumberOfRepositories: tc.numberOfRepositoriesQ1,
				RepositoryByteSizes:  tc.repositoryByteSizesQ1,
			})
			cost2 := queryAnalyzer.Cost(&QueryObject{
				Query:                queryPlan2,
				NumberOfRepositories: tc.numberOfRepositoriesQ2,
				RepositoryByteSizes:  tc.repositoryByteSizesQ2,
			})
			tc.compare(t, cost1, cost2)

		})
	}
}
