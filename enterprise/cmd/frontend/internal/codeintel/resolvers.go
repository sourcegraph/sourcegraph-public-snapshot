package codeintel

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

type Resolver struct {
	autoIndexingRootResolver autoindexinggraphql.RootResolver
	codenavResolver          codenavgraphql.RootResolver
	executorResolver         executor.Resolver
	policiesRootResolver     policiesgraphql.RootResolver
	uploadsRootResolver      uploadsgraphql.RootResolver
}

func newResolver(
	autoIndexingRootResolver autoindexinggraphql.RootResolver,
	codenavResolver codenavgraphql.RootResolver,
	executorResolver executor.Resolver,
	policiesRootResolver policiesgraphql.RootResolver,
	uploadsRootResolver uploadsgraphql.RootResolver,
) *Resolver {
	return &Resolver{
		autoIndexingRootResolver: autoIndexingRootResolver,
		codenavResolver:          codenavResolver,
		executorResolver:         executorResolver,
		policiesRootResolver:     policiesRootResolver,
		uploadsRootResolver:      uploadsRootResolver,
	}
}

func (r *Resolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"LSIFUpload": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.LSIFUploadByID(ctx, id)
		},
		"LSIFIndex": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.LSIFIndexByID(ctx, id)
		},
		"CodeIntelligenceConfigurationPolicy": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.ConfigurationPolicyByID(ctx, id)
		},
	}
}

func (r *Resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploadByID
func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFUploadResolver, err error) {
	return r.uploadsRootResolver.LSIFUploadByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploads(ctx context.Context, args *uploadsgraphql.LSIFUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.uploadsRootResolver.LSIFUploads(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *uploadsgraphql.LSIFRepositoryUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.uploadsRootResolver.LSIFUploadsByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.uploadsRootResolver.DeleteLSIFUpload(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFUploads(ctx context.Context, args *uploadsgraphql.DeleteLSIFUploadsArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.uploadsRootResolver.DeleteLSIFUploads(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexByID
func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexes(ctx context.Context, args *autoindexinggraphql.LSIFIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexes(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *autoindexinggraphql.LSIFRepositoryIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexesByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence index data
func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.DeleteLSIFIndex(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFIndexes(ctx context.Context, args *autoindexinggraphql.DeleteLSIFIndexesArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.DeleteLSIFIndexes(ctx, args)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (_ uploadsgraphql.CodeIntelligenceCommitGraphResolver, err error) {
	return r.uploadsRootResolver.CommitGraph(ctx, id)
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *autoindexinggraphql.QueueAutoIndexJobsForRepoArgs) (_ []sharedresolvers.LSIFIndexResolver, err error) {
	return r.autoIndexingRootResolver.QueueAutoIndexJobsForRepo(ctx, args)
}

func (r *Resolver) RequestLanguageSupport(ctx context.Context, args *autoindexinggraphql.RequestLanguageSupportArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.RequestLanguageSupport(ctx, args)
}

func (r *Resolver) RequestedLanguageSupport(ctx context.Context) (_ []string, err error) {
	return r.autoIndexingRootResolver.RequestedLanguageSupport(ctx)
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *codenavgraphql.GitBlobLSIFDataArgs) (_ codenavgraphql.GitBlobLSIFDataResolver, err error) {
	return r.codenavResolver.GitBlobLSIFData(ctx, args)
}

func (r *Resolver) GitBlobCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (_ autoindexinggraphql.GitBlobCodeIntelSupportResolver, err error) {
	return r.autoIndexingRootResolver.GitBlobCodeIntelInfo(ctx, args)
}

func (r *Resolver) GitTreeCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (resolver autoindexinggraphql.GitTreeCodeIntelSupportResolver, err error) {
	return r.autoIndexingRootResolver.GitTreeCodeIntelInfo(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicyByID
func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.policiesRootResolver.ConfigurationPolicyByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *policiesgraphql.CodeIntelligenceConfigurationPoliciesArgs) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	return r.policiesRootResolver.CodeIntelligenceConfigurationPolicies(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.CreateCodeIntelligenceConfigurationPolicyArgs) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.policiesRootResolver.CreateCodeIntelligenceConfigurationPolicy(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.policiesRootResolver.UpdateCodeIntelligenceConfigurationPolicy(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.policiesRootResolver.DeleteCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *Resolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error) {
	return r.autoIndexingRootResolver.RepositorySummary(ctx, id)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ autoindexinggraphql.IndexConfigurationResolver, err error) {
	return r.autoIndexingRootResolver.IndexConfiguration(ctx, id)
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *autoindexinggraphql.UpdateRepositoryIndexConfigurationArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.UpdateRepositoryIndexConfiguration(ctx, args)
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, args *policiesgraphql.PreviewRepositoryFilterArgs) (_ policiesgraphql.RepositoryFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewRepositoryFilter(ctx, args)
}

func (r *Resolver) CodeIntelligenceInferenceScript(ctx context.Context) (_ string, err error) {
	return r.autoIndexingRootResolver.CodeIntelligenceInferenceScript(ctx)
}

func (r *Resolver) UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *autoindexinggraphql.UpdateCodeIntelligenceInferenceScriptArgs) (_ *gql.EmptyResponse, err error) {
	return &gql.EmptyResponse{}, r.autoIndexingRootResolver.UpdateCodeIntelligenceInferenceScript(ctx, args)
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *policiesgraphql.PreviewGitObjectFilterArgs) (_ []policiesgraphql.GitObjectFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewGitObjectFilter(ctx, id, args)
}
