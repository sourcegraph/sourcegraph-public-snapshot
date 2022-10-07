package priority

import "strings"

// This package works by associating a cost to a search query according to a number of dimensions:
// - How expensive the search query might be, based on its type
// - How precise it is

// The query analyzer does not deal with how a search query should be prioritized according to its cost.

type QueryAnalyzer struct {
	QueryObject

	costHandlers []CostHeuristicFunc
}

type QueryObject struct {
	query string
	// the object can be augmented with repository information, or anything else.
}

type CostHeuristicFunc struct {
	fn     func(QueryObject) int
	weight int
}

var DefaultCostHandlers = []CostHeuristicFunc{
	{queryTypeHeuristic, 10},
}

func NewQueryAnalyzer(object QueryObject, handlers ...CostHeuristicFunc) *QueryAnalyzer {
	qa := &QueryAnalyzer{
		QueryObject: object,
	}
	for _, h := range handlers {
		qa.costHandlers = append(qa.costHandlers, h)
	}
	return qa
}

func (a *QueryAnalyzer) Cost() int {
	totalCost := 0
	for _, handler := range a.costHandlers {
		totalCost += handler.fn(a.QueryObject) * handler.weight
	}
	return totalCost
}

func queryTypeHeuristic(o QueryObject) int {
	// simplified
	if strings.Contains(o.query, "patterntype:structural") {
		return 100
	}
	return 0
}
