package priority

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
)

func TestQueryAnalyzerCost(t *testing.T) {
	defaultHandlers := []CostHeuristic{QueryCost}

	testCases := []struct {
		name        string
		query       string
		handlers    []CostHeuristic
		higherThan  float64
		smallerThan float64
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryAnalyzer := NewQueryAnalyzer(tc.handlers...)
			queryPlan, err := querybuilder.ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			cost := queryAnalyzer.Cost(QueryObject{Query: queryPlan})
			if cost < tc.higherThan {
				t.Errorf("expected cost to be higher than %f, got %f", tc.higherThan, cost)
			}
			if cost > tc.smallerThan {
				t.Errorf("expected cost to be smaller than %f, got %f", tc.smallerThan, cost)
			}
		})
	}
}
