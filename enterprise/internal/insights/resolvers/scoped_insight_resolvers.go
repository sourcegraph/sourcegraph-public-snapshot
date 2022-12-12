package resolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.ScopedInsightQueryPayloadResolver = &scopedInsightQueryPreviewResolver{}

func (r *Resolver) ValidateScopedInsightQuery(ctx context.Context, args graphqlbackend.ValidateScopedInsightQueryArgs) (graphqlbackend.ScopedInsightQueryPayloadResolver, error) {
	plan, err := querybuilder.ParseQuery(args.Input.Query, "standard")
	if err != nil {
		return nil, errors.Wrap(err, "the input query is invalid")
	}
	if reason, invalid := isValidScopeQuery(plan); !invalid {
		return nil, errors.New(reason)
	}
	// todo make new query
	// todo run query match number
	return &scopedInsightQueryPreviewResolver{query: args.Input.Query}, nil
}

// Possible reasons that a scope query is invalid.
const containsPattern = "the query cannot be used for scoping because it contains a pattern: `%s`."
const containsDisallowedFilter = "the query cannot be used for scoping because it contains a disallowed filter: `%s`."

// isValidScopeQuery takes a query plan and returns whether the query is a valid scope query, that is it only contains
// repo filters or boolean predicates.
func isValidScopeQuery(plan query.Plan) (string, bool) {
	for _, basic := range plan {
		if basic.Pattern != nil {
			return fmt.Sprintf(containsPattern, basic.PatternString()), false
		}
		for _, parameter := range basic.Parameters {
			field := strings.ToLower(parameter.Field)
			// Only allowed filters are: repo, fork, archived, count.
			if field != query.FieldRepo && field != query.FieldFork && field != query.FieldArchived && field != query.FieldCount {
				return fmt.Sprintf(containsDisallowedFilter, parameter.Field), false
			}
		}
	}
	return "", true
}

type scopedInsightQueryPreviewResolver struct {
	numberOfRepositories int32
	query                string
}

func (s *scopedInsightQueryPreviewResolver) NumberOfRepositories(ctx context.Context) int32 {
	return s.numberOfRepositories
}

func (s *scopedInsightQueryPreviewResolver) Query(ctx context.Context) string {
	return s.query
}
