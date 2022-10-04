package graphql

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API. This
// resolver concerns itself with GraphQL/API-specific behaviors (auth, validation, marshaling, etc.).
// All code intel-specific behavior is delegated to the underlying resolver instance, which is defined
// in the parent package.
type Resolver struct {
	resolver resolvers.Resolver
}

// NewResolver creates a new Resolver with the given resolver that defines all code intel-specific behavior.
func NewResolver(resolver resolvers.Resolver) gql.CodeIntelResolver {
	baseResolver := &Resolver{
		resolver: resolver,
	}

	return baseResolver
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
	return r.resolver.ExecutorResolver()
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploadByID
func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFUploadResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploadByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploads(ctx context.Context, args *uploadsgraphql.LSIFUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploads(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetUploads
func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *uploadsgraphql.LSIFRepositoryUploadsQueryArgs) (_ sharedresolvers.LSIFUploadConnectionResolver, err error) {
	return r.resolver.UploadRootResolver().LSIFUploadsByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.UploadRootResolver().DeleteLSIFUpload(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence upload data
func (r *Resolver) DeleteLSIFUploads(ctx context.Context, args *uploadsgraphql.DeleteLSIFUploadsArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.UploadRootResolver().DeleteLSIFUploads(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexByID
func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ sharedresolvers.LSIFIndexResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexes(ctx context.Context, args *autoindexinggraphql.LSIFIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexes(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetIndexes
func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *autoindexinggraphql.LSIFRepositoryIndexesQueryArgs) (_ sharedresolvers.LSIFIndexConnectionResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().LSIFIndexesByRepo(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence index data
func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.AutoIndexingRootResolver().DeleteLSIFIndex(ctx, args)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (_ uploadsgraphql.CodeIntelligenceCommitGraphResolver, err error) {
	return r.resolver.UploadRootResolver().CommitGraph(ctx, id)
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *autoindexinggraphql.QueueAutoIndexJobsForRepoArgs) (_ []sharedresolvers.LSIFIndexResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().QueueAutoIndexJobsForRepo(ctx, args)
}

func (r *Resolver) RequestLanguageSupport(ctx context.Context, args *autoindexinggraphql.RequestLanguageSupportArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.AutoIndexingRootResolver().RequestLanguageSupport(ctx, args)
}

func (r *Resolver) RequestedLanguageSupport(ctx context.Context) (_ []string, err error) {
	return r.resolver.AutoIndexingRootResolver().RequestedLanguageSupport(ctx)
}

// 🚨 SECURITY: dbstore layer handles authz for query resolution
func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *codenavgraphql.GitBlobLSIFDataArgs) (_ codenavgraphql.GitBlobLSIFDataResolver, err error) {
	return r.resolver.CodeNavResolver().GitBlobLSIFData(ctx, args)
}

func (r *Resolver) GitBlobCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (_ autoindexinggraphql.GitBlobCodeIntelSupportResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().GitBlobCodeIntelInfo(ctx, args)
}

func (r *Resolver) GitTreeCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (resolver autoindexinggraphql.GitTreeCodeIntelSupportResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().GitTreeCodeIntelInfo(ctx, args)
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicyByID
func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.resolver.PoliciesRootResolver().ConfigurationPolicyByID(ctx, id)
}

// 🚨 SECURITY: dbstore layer handles authz for GetConfigurationPolicies
func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *policiesgraphql.CodeIntelligenceConfigurationPoliciesArgs) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	return r.resolver.PoliciesRootResolver().CodeIntelligenceConfigurationPolicies(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.CreateCodeIntelligenceConfigurationPolicyArgs) (_ policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.resolver.PoliciesRootResolver().CreateCodeIntelligenceConfigurationPolicy(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.PoliciesRootResolver().UpdateCodeIntelligenceConfigurationPolicy(ctx, args)
}

// 🚨 SECURITY: Only site admins may modify code intelligence configuration policies
func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.PoliciesRootResolver().DeleteCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *Resolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ sharedresolvers.CodeIntelRepositorySummaryResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().RepositorySummary(ctx, id)
}

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ autoindexinggraphql.IndexConfigurationResolver, err error) {
	return r.resolver.AutoIndexingRootResolver().IndexConfiguration(ctx, id)
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *autoindexinggraphql.UpdateRepositoryIndexConfigurationArgs) (_ *sharedresolvers.EmptyResponse, err error) {
	return r.resolver.AutoIndexingRootResolver().UpdateRepositoryIndexConfiguration(ctx, args)
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, args *policiesgraphql.PreviewRepositoryFilterArgs) (_ policiesgraphql.RepositoryFilterPreviewResolver, err error) {
	return r.resolver.PoliciesRootResolver().PreviewRepositoryFilter(ctx, args)
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *policiesgraphql.PreviewGitObjectFilterArgs) (_ []policiesgraphql.GitObjectFilterPreviewResolver, err error) {
	return r.resolver.PoliciesRootResolver().PreviewGitObjectFilter(ctx, id, args)
}
