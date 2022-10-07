package priority

import "strings"

// The query analyzer gives a cost to a search query according to a number of dimensions.
// It does not deal with how a search query should be prioritized according to its cost.

type QueryAnalyzer struct {
	QueryObject

	costHandlers []CostHeuristic
}

type QueryObject struct {
	query string
	// the object can be augmented with repository information, or anything else.
}

type CostHeuristic struct {
	fn     func(QueryObject) int
	weight int
}

var DefaultCostHandlers = []CostHeuristic{
	{queryTypeCost, 10},
}

func NewQueryAnalyzer(object QueryObject, handlers []CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		QueryObject:  object,
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost() int {
	totalCost := 0
	for _, handler := range a.costHandlers {
		totalCost += handler.fn(a.QueryObject) * handler.weight
	}
	return totalCost
}

func queryTypeCost(o QueryObject) int {
	// todo implement actual functionality
	if strings.Contains(o.query, "patterntype:structural") {
		return 100
	}
	return 0
}
