package querybuilder

import (
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func searchTypeFromString(pt string) query.SearchType {
	var searchType query.SearchType
	switch pt {
	case "literal":
		searchType = query.SearchTypeLiteral
	case "structural":
		searchType = query.SearchTypeStructural
	case "regexp", "regex":
		searchType = query.SearchTypeRegex
	case "standard":
		searchType = query.SearchTypeStandard
	case "lucky":
		searchType = query.SearchTypeLucky
	default:
		searchType = query.SearchTypeLiteral
	}
	return searchType
}

func ParseAndValidateQuery(q string, patternType string) (query.Plan, error) {
	plan, err := query.Pipeline(query.Init(q, searchTypeFromString(patternType)))
	if err != nil {
		return nil, errors.Wrapf(err, "Pipeline")
	}
	return plan, nil
}

// ParametersFromQueryPlan expects a valid query plan and returns all parameters from it, e.g. context:global.
func ParametersFromQueryPlan(plan query.Plan) []query.Parameter {
	var parameters []query.Parameter
	for _, basic := range plan {
		parameters = append(parameters, basic.Parameters...)
	}
	return parameters
}
