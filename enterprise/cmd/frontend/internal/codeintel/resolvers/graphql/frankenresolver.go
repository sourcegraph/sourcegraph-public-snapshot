package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
)

type frankenResolver struct {
	*Resolver
	gql.AutoindexingServiceResolver
	gql.UploadsServiceResolver
	gql.PoliciesServiceResolver
}

func (r *frankenResolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error) {
	return r.Resolver.LSIFIndexByID(ctx, id)
}

func (r *frankenResolver) LSIFIndexes(ctx context.Context, args *autoindexinggraphql.LSIFIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.Resolver.LSIFIndexes(ctx, args)
}

func (r *frankenResolver) LSIFIndexesByRepo(ctx context.Context, args *autoindexinggraphql.LSIFRepositoryIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.Resolver.LSIFIndexesByRepo(ctx, args)
}

func (r *frankenResolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.Resolver.DeleteLSIFIndex(ctx, args)
}

func (r *frankenResolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *autoindexinggraphql.QueueAutoIndexJobsForRepoArgs) (_ []sharedresolvers.LSIFIndexResolver, err error) {
	return r.Resolver.QueueAutoIndexJobsForRepo(ctx, args)
}

func (r *frankenResolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ gql.IndexConfigurationResolver, err error) {
	return r.Resolver.IndexConfiguration(ctx, id)
}

func (r *frankenResolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *autoindexinggraphql.UpdateRepositoryIndexConfigurationArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.Resolver.UpdateRepositoryIndexConfiguration(ctx, args)
}

func (r *frankenResolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error) {
	return r.Resolver.RepositorySummary(ctx, id)
}

func (r *frankenResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFUploadResolver, err error) {
	return r.Resolver.LSIFUploadByID(ctx, id)
}

func (r *frankenResolver) LSIFUploads(ctx context.Context, args *uploadsgraphql.LSIFUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.Resolver.LSIFUploads(ctx, args)
}

func (r *frankenResolver) LSIFUploadsByRepo(ctx context.Context, args *uploadsgraphql.LSIFRepositoryUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.Resolver.LSIFUploadsByRepo(ctx, args)
}

func (r *frankenResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.Resolver.DeleteLSIFUpload(ctx, args)
}

func (r *frankenResolver) CommitGraph(ctx context.Context, id graphql.ID) (_ uploadsgraphql.CodeIntelligenceCommitGraphResolver, err error) {
	return r.Resolver.CommitGraph(ctx, id)
}

func (r *frankenResolver) getPoliciesServiceResolver() gql.PoliciesServiceResolver {
	return r.Resolver
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
