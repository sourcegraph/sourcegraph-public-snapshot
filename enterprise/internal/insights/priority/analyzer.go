package priority

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// The query analyzer gives a cost to a search query according to a number of dimensions.
// It does not deal with how a search query should be prioritized according to its cost.

type QueryAnalyzer struct {
	costHandlers []CostHeuristic
}

type QueryObject struct {
	query query.Plan
	// the object can be augmented with repository information, or anything else.
}

type CostHeuristic struct {
	fn     func(QueryObject) int
	weight int
}

var DefaultCostHandlers = []CostHeuristic{
	{queryContentCost, 10},
}

func NewQueryAnalyzer(object QueryObject, handlers []CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost(o QueryObject) int {
	totalCost := 0
	for _, handler := range a.costHandlers {
		totalCost += handler.fn(o) * handler.weight
	}
	return totalCost
}

// analyze a query according to:
// - the kind of content it will match (e.g. structural, literal)
// - how precise the content is (e.g. file: selector)

func queryContentCost(o QueryObject) int {
	var contentCost int
	nodes := o.query.ToQ()
	queryString := nodes.String()
	searchType, _ := querybuilder.DetectSearchType(queryString, "structural")
	if searchType == query.SearchTypeStructural {
		contentCost += 1000
	}
	if searchType == query.SearchTypeRegex {
		// todo detect if capture group pattern would match loads
		// (although, if that is the case, do we even want to allow such a query?)
		contentCost += 800
	}

	var unindexed, diff, commit bool
	// todo visit each parameter and:
	// if unindexed, diff, commit, set (slow stuff: matches slow)
	query.VisitParameter(nodes, func(field, value string, negated bool, annotation query.Annotation) {
		if field == "index" && (value == "no" || value == "n") {
			unindexed = true
		}
		if field == "type" {
			if value == "diff" {
				diff = true
			} else if value == "commit" {
				commit = true
			}
		}
	})
	if unindexed {
		contentCost += 1000
	}
	if diff {
		contentCost += 1000
	}
	if commit {
		contentCost += 800
	}

	return contentCost
}

func queryPrecisionCost(o QueryObject) int {
	return 0
}
