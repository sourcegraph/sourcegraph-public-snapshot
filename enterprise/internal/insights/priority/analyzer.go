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
	Query query.Plan
	// the object can be augmented with repository information, or anything else of value.
}

type CostHeuristic func(QueryObject) float64

func NewQueryAnalyzer(handlers ...CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost(o QueryObject) float64 {
	var totalCost float64
	for _, handler := range a.costHandlers {
		totalCost += handler(o)
	}
	if totalCost < 0 {
		return 0
	}
	return totalCost
}

func QueryCost(o QueryObject) float64 {
	var cost float64
	for _, basic := range o.Query {
		if basic.IsStructural() {
			cost += StructuralCost
		} else if basic.IsRegexp() {
			cost += RegexpCost
		} else {
			cost += LiteralCost
		}
	}

	var diff, commit bool
	query.VisitParameter(o.Query.ToQ(), func(field, value string, negated bool, annotation query.Annotation) {
		if field == "type" {
			if value == "diff" {
				diff = true
			} else if value == "commit" {
				commit = true
			}
		}
	})
	if diff {
		cost *= DiffMultiplier
	}
	if commit {
		cost *= CommitMultiplier
	}

	parameters := querybuilder.ParametersFromQueryPlan(o.Query)
	if parameters.Index() == query.No {
		cost *= UnindexedMultiplier
	}
	if parameters.Exists(query.FieldAuthor) {
		cost *= AuthorMultiplier
	}
	if parameters.Exists(query.FieldFile) {
		cost *= FileMultiplier
	}
	if parameters.Exists(query.FieldLang) {
		cost *= LangMultiplier
	}

	archived := parameters.Archived()
	if archived != nil {
		if *archived == query.Yes {
			cost *= YesMultiplier
		} else if *archived == query.Only {
			cost *= OnlyMultiplier
		}
	}
	fork := parameters.Fork()
	if fork != nil && (*fork == query.Yes || *fork == query.Only) {
		if *fork == query.Yes {
			cost *= YesMultiplier
		} else if *fork == query.Only {
			cost *= OnlyMultiplier
		}
	}
	return cost
}
