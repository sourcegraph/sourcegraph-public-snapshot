package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
)

var (
	_ graphqlbackend.ScopedInsightQueryPayloadResultResolver       = &scopedInsightQueryResultResolver{}
	_ graphqlbackend.ScopedInsightQueryPayloadResolver             = &scopedInsightQueryPayloadResolver{}
	_ graphqlbackend.ScopedInsightQueryPayloadNotAvailableResolver = &scopedInsightQueryPayloadNotAvailableResolver{}
)

func (r *Resolver) ValidateScopedInsightQuery(ctx context.Context, args graphqlbackend.ValidateScopedInsightQueryArgs) (graphqlbackend.ScopedInsightQueryPayloadResultResolver, error) {
	plan, err := querybuilder.ParseQuery(args.Input.Query, "literal")
	if err != nil {
		return &scopedInsightQueryResultResolver{
			resolver: &scopedInsightQueryPayloadNotAvailableResolver{
				reason:     fmt.Sprintf("the input query is invalid: %v", err),
				reasonType: types.INVALID_SCOPE_QUERY,
			},
		}, nil
	}
	if reason, invalid := isValidScopeQuery(plan); !invalid {
		return &scopedInsightQueryResultResolver{
			resolver: &scopedInsightQueryPayloadNotAvailableResolver{reason: reason, reasonType: types.UNSUPPORTED_SEARCH_ARGUMENT},
		}, nil
	}

	var numberOfRepositories *int32
	if args.Input.FetchNumberOfRepositories {
		executor := query.NewStreamingExecutor(r.postgresDB, time.Now)
		repos, err := executor.ExecuteRepoList(ctx, args.Input.Query)
		if err != nil {
			return &scopedInsightQueryResultResolver{
				resolver: &scopedInsightQueryPayloadNotAvailableResolver{
					reason:     fmt.Sprintf("executing the repository search errored: %v", err),
					reasonType: types.SCOPE_SEARCH_ERROR,
				},
			}, nil
		}
		number := int32(len(repos))
		numberOfRepositories = &number
	}

	return &scopedInsightQueryResultResolver{
		resolver: &scopedInsightQueryPayloadResolver{query: args.Input.Query, numberOfRepositories: numberOfRepositories},
	}, nil
}

// Possible reasons that a scope query is invalid.
const containsPattern = "the query cannot be used for scoping because it contains a pattern: `%s`."
const containsDisallowedFilter = "the query cannot be used for scoping because it contains a disallowed filter: `%s`."

// isValidScopeQuery takes a query plan and returns whether the query is a valid scope query, that is it only contains
// repo filters or boolean predicates.
func isValidScopeQuery(plan searchquery.Plan) (string, bool) {
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
		}
	}
	return "", true
}

type scopedInsightQueryPayloadResolver struct {
	numberOfRepositories *int32
	query                string
}

func (s *scopedInsightQueryPayloadResolver) NumberOfRepositories(ctx context.Context) *int32 {
	return s.numberOfRepositories
}

func (s *scopedInsightQueryPayloadResolver) Query(ctx context.Context) string {
	return s.query
}

type scopedInsightQueryPayloadNotAvailableResolver struct {
	reason     string
	reasonType types.ScopedInsightQueryPayloadNotAvailableReasonType
}

func (r *scopedInsightQueryPayloadNotAvailableResolver) Reason() string {
	return r.reason
}

func (r *scopedInsightQueryPayloadNotAvailableResolver) ReasonType() string {
	return string(r.reasonType)
}

type scopedInsightQueryResultResolver struct {
	resolver any
}

func (r *scopedInsightQueryResultResolver) ToScopedInsightQueryPayload() (graphqlbackend.ScopedInsightQueryPayloadResolver, bool) {
	res, ok := r.resolver.(*scopedInsightQueryPayloadResolver)
	return res, ok
}

func (r *scopedInsightQueryResultResolver) ToScopedInsightQueryPayloadNotAvailable() (graphqlbackend.ScopedInsightQueryPayloadNotAvailableResolver, bool) {
	res, ok := r.resolver.(*scopedInsightQueryPayloadNotAvailableResolver)
	return res, ok
}
