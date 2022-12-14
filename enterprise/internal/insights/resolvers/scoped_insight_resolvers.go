package resolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
)

var (
	_ graphqlbackend.ScopedInsightQueryPayloadResolver = &scopedInsightQueryPayloadResolver{}
	_ graphqlbackend.RepositoryPreviewPayloadResolver  = &repositorityPreviewPayloadResolver{}
)

func (r *Resolver) ValidateScopedInsightQuery(ctx context.Context, args graphqlbackend.ValidateScopedInsightQueryArgs) (graphqlbackend.ScopedInsightQueryPayloadResolver, error) {
	plan, err := querybuilder.ParseQuery(args.Query, "literal")
	if err != nil {
		invalidReason := fmt.Sprintf("the input query is invalid: %v", err)
		return &scopedInsightQueryPayloadResolver{
			query:         args.Query,
			isValid:       false,
			invalidReason: &invalidReason,
		}, nil
	}
	if reason, invalid := isValidScopeQuery(plan); !invalid {
		return &scopedInsightQueryPayloadResolver{
			query:         args.Query,
			isValid:       false,
			invalidReason: &reason,
		}, nil
	}
	return &scopedInsightQueryPayloadResolver{
		query:   args.Query,
		isValid: true,
	}, nil
}

//var numberOfRepositories *int32
//if args.Input.FetchNumberOfRepositories {
//executor := query.NewStreamingExecutor(r.postgresDB, time.Now)
//repos, err := executor.ExecuteRepoList(ctx, args.Input.Query)
//if err != nil {
//return &scopedInsightQueryResultResolver{
//resolver: &scopedInsightQueryPayloadNotAvailableResolver{
//reason:     fmt.Sprintf("executing the repository search errored: %v", err),
//reasonType: types.SCOPE_SEARCH_ERROR,
//},
//}, nil
//}
//number := int32(len(repos))
//numberOfRepositories = &number
//}
//
//return &scopedInsightQueryResultResolver{
//resolver: &scopedInsightQueryPayloadResolver{query: args.Input.Query, numberOfRepositories: numberOfRepositories},
//}, nil

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

func (r *repositorityPreviewPayloadResolver) PreviewRepositoriesFromQuery(ctx context.Context, args graphqlbackend.PreviewRepositoriesFromQueryArgs) (graphqlbackend.RepositoryPreviewPayloadResolver, error) {
	return nil, nil
}

type scopedInsightQueryPayloadResolver struct {
	query         string
	isValid       bool
	invalidReason *string
}

func (r *scopedInsightQueryPayloadResolver) Query(ctx context.Context) string {
	return r.query
}

func (r *scopedInsightQueryPayloadResolver) IsValid(ctx context.Context) bool {
	return r.isValid
}

func (r *scopedInsightQueryPayloadResolver) InvalidReason(ctx context.Context) *string {
	return r.invalidReason
}

type repositorityPreviewPayloadResolver struct {
	query                string
	numberOfRepositories *int32
}

func (r *repositorityPreviewPayloadResolver) Query(ctx context.Context) string {
	return r.query
}

func (r *repositorityPreviewPayloadResolver) NumberOfRepositories(ctx context.Context) *int32 {
	return r.numberOfRepositories
}
