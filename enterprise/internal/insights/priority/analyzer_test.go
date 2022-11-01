package priority

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
)

const (
	Simple        float64 = 50
	Slow          float64 = 500
	Long          float64 = 5000
	LikelyTimeout float64 = 10000
) // values that could associate a speed to a floating point

func TestQueryAnalyzerCost(t *testing.T) {
	defaultHandlers := []CostHeuristic{QueryCost, RepositoriesCost}

	testCases := []struct {
		name                 string
		query                string
		numberOfRepositories int64
		repositoryByteSizes  []int64
		handlers             []CostHeuristic
		higherThan           float64
		smallerThan          float64
	}{
		{
			name:        "literal diff query should be magnitudes higher",
			query:       "Type:diff insights",
			handlers:    defaultHandlers,
			higherThan:  Long,
			smallerThan: Long * 1000,
		},
		{
			name:        "literal diff query with author should reduce complexity",
			query:       "type:diff author:someone insights",
			handlers:    defaultHandlers,
			higherThan:  Slow,
			smallerThan: Long,
		},
		{
			name:        "regexp query with reduced complexity not slow",
			query:       "patterntype:regexp [0-9]+ lang:go",
			handlers:    defaultHandlers,
			higherThan:  Simple,
			smallerThan: Slow,
		},
		{
			name:        "very specific query super speedy",
			query:       "file:insights lang:go DashboardResolver",
			handlers:    defaultHandlers,
			higherThan:  0.0,
			smallerThan: Simple,
		},
		{
			name:                 "regexp query with reduced complexity but many repos slow",
			query:                "patterntype:regexp [0-9]+ lang:go",
			numberOfRepositories: 30000,
			handlers:             defaultHandlers,
			higherThan:           Slow,
			smallerThan:          Long,
		},
		{
			name:                "literal query on gigarepo is long but will complete",
			query:               "patterntype:literal context.Context",
			repositoryByteSizes: []int64{gigarepoSizethreshold},
			handlers:            defaultHandlers,
			higherThan:          Long,
			smallerThan:         LikelyTimeout,
		},
		{
			name:                 "total annihilation query",
			query:                "patterntype:structural [a] archive:yes fork:yes index:no",
			numberOfRepositories: 3,
			repositoryByteSizes:  []int64{100, megarepoSizeThresold, gigarepoSizethreshold * 2},
			handlers:             defaultHandlers,
			higherThan:           LikelyTimeout * 10000,
			smallerThan:          LikelyTimeout * 100000,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryAnalyzer := NewQueryAnalyzer(tc.handlers...)
			queryPlan, err := querybuilder.ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			cost := queryAnalyzer.Cost(&QueryObject{
				Query:                queryPlan,
				NumberOfRepositories: tc.numberOfRepositories,
				RepositoryByteSizes:  tc.repositoryByteSizes,
			})
			if cost < tc.higherThan {
				t.Errorf("expected cost to be higher than %f, got %f", tc.higherThan, cost)
			}
			if cost > tc.smallerThan {
				t.Errorf("expected cost to be smaller than %f, got %f", tc.smallerThan, cost)
			}
		})
	}
}
