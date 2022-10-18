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
	Query                query.Plan
	NumberOfRepositories int64
	RepositoryByteSizes  []int64 // size of repositories in bytes, if known

	Cost float64
}

type CostHeuristic func(*QueryObject)

func NewQueryAnalyzer(handlers ...CostHeuristic) *QueryAnalyzer {
	return &QueryAnalyzer{
		costHandlers: handlers,
	}
}

func (a *QueryAnalyzer) Cost(o *QueryObject) float64 {
	for _, handler := range a.costHandlers {
		handler(o)
	}
	if o.Cost < 0.0 {
		return 0.0
	}
	return o.Cost
}

func QueryCost(o *QueryObject) {
	for _, basic := range o.Query {
		if basic.IsStructural() {
			o.Cost += StructuralCost
		} else if basic.IsRegexp() {
			o.Cost += RegexpCost
		} else {
			o.Cost += LiteralCost
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
		o.Cost *= DiffMultiplier
	}
	if commit {
		o.Cost *= CommitMultiplier
	}

	parameters := querybuilder.ParametersFromQueryPlan(o.Query)
	if parameters.Index() == query.No {
		o.Cost *= UnindexedMultiplier
	}
	if parameters.Exists(query.FieldAuthor) {
		o.Cost *= AuthorMultiplier
	}
	if parameters.Exists(query.FieldFile) {
		o.Cost *= FileMultiplier
	}
	if parameters.Exists(query.FieldLang) {
		o.Cost *= LangMultiplier
	}

	archived := parameters.Archived()
	if archived != nil {
		if *archived == query.Yes {
			o.Cost *= YesMultiplier
		} else if *archived == query.Only {
			o.Cost *= OnlyMultiplier
		}
	}
	fork := parameters.Fork()
	if fork != nil && (*fork == query.Yes || *fork == query.Only) {
		if *fork == query.Yes {
			o.Cost *= YesMultiplier
		} else if *fork == query.Only {
			o.Cost *= OnlyMultiplier
		}
	}
}

var (
	megarepoSizeThresold  int64 = 5368709120                // 5GB
	gigarepoSizethreshold int64 = megarepoSizeThresold * 10 // 50GB
)

func RepositoriesCost(o *QueryObject) {
	if o.Cost <= 0.0 {
		o.Cost = 1 // if this handler is called on its own we still want it to impact the cost.
	}

	if o.NumberOfRepositories > 10000 {
		o.Cost *= ManyRepositoriesMultiplier
	}

	var megarepo, gigarepo bool
	for _, byteSize := range o.RepositoryByteSizes {
		if byteSize >= gigarepoSizethreshold {
			gigarepo = true
		}
		if byteSize >= megarepoSizeThresold {
			megarepo = true
		}
	}
	if gigarepo {
		o.Cost *= GigarepoMultiplier
	} else if megarepo {
		o.Cost *= MegarepoMultiplier
	}
}
