package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	if reason, invalid := querybuilder.IsValidScopeQuery(plan); !invalid {
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

func (r *Resolver) PreviewRepositoriesFromQuery(ctx context.Context, args graphqlbackend.PreviewRepositoriesFromQueryArgs) (graphqlbackend.RepositoryPreviewPayloadResolver, error) {
	plan, err := querybuilder.ParseQuery(args.Query, "literal")
	if err != nil {
		return nil, errors.Wrap(err, "the input query is invalid")
	}
	if reason, invalid := querybuilder.IsValidScopeQuery(plan); !invalid {
		return nil, errors.Newf("the input query cannot be used for previewing repositories: %v", reason)
	}

	repoScopeQuery, err := querybuilder.RepositoryScopeQuery(args.Query)
	if err != nil {
		return nil, errors.Wrap(err, "could not build repository scope query")
	}

	executor := query.NewStreamingRepoQueryExecutor(r.logger.Scoped("StreamingRepoQueryExecutor"))
	repos, err := executor.ExecuteRepoList(ctx, repoScopeQuery.String())
	if err != nil {
		return nil, errors.Wrap(err, "executing the repository search errored")
	}
	number := int32(len(repos))

	return &repositorityPreviewPayloadResolver{
		query:                repoScopeQuery.String(),
		numberOfRepositories: &number,
	}, nil
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
