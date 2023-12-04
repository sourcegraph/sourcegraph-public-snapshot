package priority

import (
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// The query analyzer gives a cost to a search query according to a number of heuristics.
// It does not deal with how a search query should be prioritized according to its cost.

type QueryAnalyzer struct {
	costHandlers []CostHeuristic
}

type QueryObject struct {
	Query                query.Plan
	NumberOfRepositories int64
	RepositoryByteSizes  []int64 // size of repositories in bytes, if known

	cost float64
}

type CostHeuristic func(*QueryObject)

func DefaultQueryAnalyzer() *QueryAnalyzer {
	return NewQueryAnalyzer(QueryCost, RepositoriesCost)
}

func NewQueryAnalyzer(handlers ...CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost(o *QueryObject) float64 {
	for _, handler := range a.costHandlers {
		handler(o)
	}
	if o.cost < 0.0 {
		return 0.0
	}
	return o.cost
}

func QueryCost(o *QueryObject) {
	for _, basic := range o.Query {
		if basic.IsStructural() {
			o.cost += StructuralCost
		} else if basic.IsRegexp() {
			o.cost += RegexpCost
		} else {
			o.cost += LiteralCost
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
		o.cost *= DiffMultiplier
	}
	if commit {
		o.cost *= CommitMultiplier
	}

	parameters := querybuilder.ParametersFromQueryPlan(o.Query)
	if parameters.Index() == query.No {
		o.cost *= UnindexedMultiplier
	}
	if parameters.Exists(query.FieldAuthor) {
		o.cost *= AuthorMultiplier
	}
	if parameters.Exists(query.FieldFile) {
		o.cost *= FileMultiplier
	}
	if parameters.Exists(query.FieldLang) {
		o.cost *= LangMultiplier
	}

	archived := parameters.Archived()
	if archived != nil {
		if *archived == query.Yes {
			o.cost *= YesMultiplier
		} else if *archived == query.Only {
			o.cost *= OnlyMultiplier
		}
	}
	fork := parameters.Fork()
	if fork != nil && (*fork == query.Yes || *fork == query.Only) {
		if *fork == query.Yes {
			o.cost *= YesMultiplier
		} else if *fork == query.Only {
			o.cost *= OnlyMultiplier
		}
	}
}

var (
	megarepoSizeThreshold int64 = 5368709120                 // 5GB
	gigarepoSizeThreshold       = megarepoSizeThreshold * 10 // 50GB
)

func RepositoriesCost(o *QueryObject) {
	if o.cost <= 0.0 {
		o.cost = 1 // if this handler is called on its own we still want it to impact the cost.
	}

	if o.NumberOfRepositories > 100 {
		o.cost *= float64(o.NumberOfRepositories) / 100.0
	}

	var megarepo, gigarepo bool
	for _, byteSize := range o.RepositoryByteSizes {
		if byteSize >= gigarepoSizeThreshold {
			gigarepo = true
		}
		if byteSize >= megarepoSizeThreshold {
			megarepo = true
		}
	}
	if gigarepo {
		o.cost *= GigarepoMultiplier
	} else if megarepo {
		o.cost *= MegarepoMultiplier
	}
}
