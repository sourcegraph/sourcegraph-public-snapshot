package priority

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// The query analyzer gives a cost to a search query according to a number of heuristics.
// It does not deal with how a search query should be prioritized according to its cost.

type QueryAnalyzer struct {
	costHandlers []CostHeuristic
}

type QueryObject struct {
	query query.Plan
	// the object can be augmented with repository information, or anything else of value.
}

type CostHeuristic struct {
	fn     func(QueryObject) int
	weight int
}

var DefaultCostHandlers = []CostHeuristic{
	{queryContentCost, 1},
	{queryScopeCost, 2},
}

func NewQueryAnalyzer(handlers []CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost(o QueryObject) int {
	totalCost := 0
	for _, handler := range a.costHandlers {
		totalCost += handler.fn(o) * handler.weight
	}
	if totalCost < 0 {
		return 0
	}
	return totalCost
}

// queryContentCost will derive a cost based on the kind of content a query would match.
func queryContentCost(o QueryObject) int {
	var contentCost int
	for _, basic := range o.query {
		if basic.IsStructural() {
			contentCost += 1000
		}
		if basic.IsRegexp() {
			contentCost += 800
		}
	}

	var diff, commit bool
	query.VisitParameter(o.query.ToQ(), func(field, value string, negated bool, annotation query.Annotation) {
		if field == "type" {
			if value == "diff" {
				diff = true
			} else if value == "commit" {
				commit = true
			}
		}
	})
	if diff {
		contentCost += 1000
	}
	if commit {
		contentCost += 800
	}

	parameters := querybuilder.ParametersFromQueryPlan(o.query)
	if parameters.Index() == query.No {
		contentCost += 1000
	}

	return contentCost
}

// queryScopeCost will derive a cost based on how precise a query is (e.g. a commit query with an author field).
func queryScopeCost(o QueryObject) int {
	var scopeCost int

	parameters := querybuilder.ParametersFromQueryPlan(o.query)
	if parameters.Exists(query.FieldFile) {
		scopeCost -= 100
	}
	if parameters.Exists(query.FieldLang) {
		scopeCost -= 50
	}

	archived := parameters.Archived()
	if archived != nil {
		if *archived == query.Yes {
			scopeCost += 50
		} else if *archived == query.Only {
			scopeCost -= 50
		}
	}
	fork := parameters.Fork()
	if fork != nil && (*fork == query.Yes || *fork == query.Only) {
		if *fork == query.Yes {
			scopeCost += 50
		} else if *fork == query.Only {
			scopeCost -= 50
		}
	}

	var diffOrCommit bool
	query.VisitParameter(o.query.ToQ(), func(field, value string, negated bool, annotation query.Annotation) {
		if field == "type" {
			if value == "diff" {
				diffOrCommit = true
			} else if value == "commit" {
				diffOrCommit = true
			}
		}
	})
	if diffOrCommit && parameters.Exists(query.FieldAuthor) {
		scopeCost -= 100
	}

	return scopeCost
}
