package priority

import "strings"

// This package works by associating a cost to a search query according to a number of dimensions:
// - How expensive the search query might be, based on its type
// - How precise it is

type QueryAnalyzer struct {
	QueryObject

	costHandlers []CostHeuristicFunc
}

type QueryObject struct {
	query string
	// the object can be augmented with repository information, or anything else.
}

type CostHeuristicFunc func(QueryObject) int

// NewQueryAnalyzer will be initialized with default cost handlers and any additional handlers defined by the caller.
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
	for _, handlerFunc := range a.costHandlers {
		totalCost += handlerFunc(a.QueryObject)
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
