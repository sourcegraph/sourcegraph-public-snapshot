package querybuilder

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func DetectSearchType(rawQuery string, patternType string) (query.SearchType, error) {
	searchType, err := client.SearchTypeFromString(patternType)
	if err != nil {
		return -1, errors.Wrap(err, "client.SearchTypeFromString")
	}
	q, err := query.Parse(rawQuery, searchType)
	if err != nil {
		return -1, errors.Wrap(err, "query.Parse")
	}
	q = query.LowercaseFieldNames(q)
	query.VisitField(q, searchquery.FieldPatternType, func(value string, _ bool, _ query.Annotation) {
		if value != "" {
			searchType, err = client.SearchTypeFromString(value)
		}
	})
	return searchType, err

}

func ParseQuery(q string, patternType string) (query.Plan, error) {
	searchType, err := DetectSearchType(q, patternType)
	if err != nil {
		return nil, errors.Wrap(err, "overrideSearchType")
	}
	plan, err := query.Pipeline(query.Init(q, searchType))
	if err != nil {
		return nil, errors.Wrap(err, "query.Pipeline")
	}
	return plan, nil
}

func ParseComputeQuery(q string) (*compute.Query, error) {
	computeQuery, err := compute.Parse(q)
	if err != nil {
		return nil, errors.Wrap(err, "compute.Parse")
	}
	return computeQuery, nil
}

// ParametersFromQueryPlan expects a valid query plan and returns all parameters from it, e.g. context:global.
func ParametersFromQueryPlan(plan query.Plan) []query.Parameter {
	var parameters []query.Parameter
	for _, basic := range plan {
		parameters = append(parameters, basic.Parameters...)
	}
	return parameters
}
