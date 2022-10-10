package priority

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
)

func TestQueryAnalyzerCost(t *testing.T) {
	// cost with default handlers
	// can nullify cost associated with specific handler
	testCases := []struct {
		query    string
		handlers []CostHeuristic
		want     autogold.Value
	}{
		{
			"type:diff author:someone insights",
			DefaultCostHandlers,
			autogold.Want("cost with default handlers", 1000),
		},
		{
			"type:diff author:someone insights",
			[]CostHeuristic{{queryContentCost, 1}, {queryPrecisionCost, 0}},
			autogold.Want("nullify cost associated with heuristic", 100),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			queryAnalyzer := NewQueryAnalyzer(DefaultCostHandlers)
			queryPlan, err := querybuilder.ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			tc.want.Equal(t, queryAnalyzer.Cost(QueryObject{queryPlan}))
		})
	}
}

func Test_queryContentCost(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			"[a] patterntype:structural",
			autogold.Want("structural query cost", 1000),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			queryPlan, err := querybuilder.ParseQuery(tc.query, "literal")
			if err != nil {
				t.Fatal(err)
			}
			cost := queryContentCost(QueryObject{queryPlan})
			tc.want.Equal(t, cost)
		})
	}
}
