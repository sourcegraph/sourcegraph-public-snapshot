package querybuilder

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func ParseComputeQuery(q string, gitserverClient gitserver.Client) (*compute.Query, error) {
	computeQuery, err := compute.Parse(q)
	if err != nil {
		return nil, errors.Wrap(err, "compute.Parse")
	}
	return computeQuery, nil
}

// ParametersFromQueryPlan expects a valid query plan and returns all parameters from it, e.g. context:global.
func ParametersFromQueryPlan(plan query.Plan) query.Parameters {
	var parameters []query.Parameter
	for _, basic := range plan {
		parameters = append(parameters, basic.Parameters...)
	}
	return parameters
}

func ContainsField(rawQuery, field string) (bool, error) {
	plan, err := ParseQuery(rawQuery, "literal")
	if err != nil {
		return false, errors.Wrap(err, "ParseQuery")
	}
	for _, basic := range plan {
		if basic.Parameters.Exists(field) {
			return true, nil
		}
	}
	return false, nil
}

// Possible reasons that a scope query is invalid.
const containsPattern = "the query cannot be used for scoping because it contains a pattern: `%s`."
const containsDisallowedFilter = "the query cannot be used for scoping because it contains a disallowed filter: `%s`."
const containsDisallowedRevision = "the query cannot be used for scoping because it contains a revision."
const containsInvalidExpression = "the query cannot be used for scoping because it is not a valid regular expression."

// IsValidScopeQuery takes a query plan and returns whether the query is a valid scope query, that is it only contains
// repo filters or boolean predicates.
func IsValidScopeQuery(plan searchquery.Plan) (string, bool) {
	for _, basic := range plan {
		if basic.Pattern != nil {
			return fmt.Sprintf(containsPattern, basic.PatternString()), false
		}
		for _, parameter := range basic.Parameters {
			field := strings.ToLower(parameter.Field)
			// Only allowed filter is repo (including repo:has predicates).
			if field != searchquery.FieldRepo {
				return fmt.Sprintf(containsDisallowedFilter, parameter.Field), false
			}
			// This is a repo filter make sure no revision was specified
			repoRevs, err := query.ParseRepositoryRevisions(parameter.Value)
			if err != nil {
				// This shouldn't be possible because it should have failed earlier when parsed
				return containsInvalidExpression, false
			}
			if len(repoRevs.Revs) > 0 {
				return containsDisallowedRevision, false
			}
		}
	}

	return "", true
}
