package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type frankenResolver struct {
	*Resolver
	gql.PoliciesServiceResolver
}

func (r *frankenResolver) getPoliciesServiceResolver() gql.PoliciesServiceResolver {
	return r.Resolver

	// TODO(efritz) - https://github.com/sourcegraph/sourcegraph/issues/33376
	// return r.PoliciesServiceResolver
}

func (r *frankenResolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.getPoliciesServiceResolver().ConfigurationPolicyByID(ctx, id)
}

func (r *frankenResolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *gql.CodeIntelligenceConfigurationPoliciesArgs) (_ gql.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	return r.getPoliciesServiceResolver().CodeIntelligenceConfigurationPolicies(ctx, args)
}

func (r *frankenResolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.CreateCodeIntelligenceConfigurationPolicyArgs) (_ gql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.getPoliciesServiceResolver().CreateCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *frankenResolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	return r.getPoliciesServiceResolver().UpdateCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *frankenResolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *gql.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *gql.EmptyResponse, err error) {
	return r.getPoliciesServiceResolver().DeleteCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *frankenResolver) PreviewRepositoryFilter(ctx context.Context, args *gql.PreviewRepositoryFilterArgs) (_ gql.RepositoryFilterPreviewResolver, err error) {
	return r.getPoliciesServiceResolver().PreviewRepositoryFilter(ctx, args)
}

func (r *frankenResolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *gql.PreviewGitObjectFilterArgs) (_ []gql.GitObjectFilterPreviewResolver, err error) {
	return r.getPoliciesServiceResolver().PreviewGitObjectFilter(ctx, id, args)
}
