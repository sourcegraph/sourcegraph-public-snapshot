package codeintel

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type Resolver struct {
	autoIndexingRootResolver resolverstubs.AutoindexingServiceResolver
	codenavResolver          resolverstubs.CodeNavServiceResolver
	policiesRootResolver     resolverstubs.PoliciesServiceResolver
	uploadsRootResolver      resolverstubs.UploadsServiceResolver
}

func newResolver(
	autoIndexingRootResolver resolverstubs.AutoindexingServiceResolver,
	codenavResolver resolverstubs.CodeNavServiceResolver,
	policiesRootResolver resolverstubs.PoliciesServiceResolver,
	uploadsRootResolver resolverstubs.UploadsServiceResolver,
) *Resolver {
	return &Resolver{
		autoIndexingRootResolver: autoIndexingRootResolver,
		codenavResolver:          codenavResolver,
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
		"PreciseIndex": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.PreciseIndexByID(ctx, id)
		},
	}
}

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (_ resolverstubs.LSIFUploadResolver, err error) {
	return r.uploadsRootResolver.LSIFUploadByID(ctx, id)
}

func (r *Resolver) LSIFUploads(ctx context.Context, args *resolverstubs.LSIFUploadsQueryArgs) (_ resolverstubs.LSIFUploadConnectionResolver, err error) {
	return r.uploadsRootResolver.LSIFUploads(ctx, args)
}

func (r *Resolver) PreciseIndexes(ctx context.Context, args *resolverstubs.PreciseIndexesQueryArgs) (_ resolverstubs.PreciseIndexConnectionResolver, err error) {
	return r.autoIndexingRootResolver.PreciseIndexes(ctx, args)
}

func (r *Resolver) PreciseIndexByID(ctx context.Context, id graphql.ID) (_ resolverstubs.PreciseIndexResolver, err error) {
	return r.autoIndexingRootResolver.PreciseIndexByID(ctx, id)
}

func (r *Resolver) DeletePreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*resolverstubs.EmptyResponse, error) {
	return r.autoIndexingRootResolver.DeletePreciseIndex(ctx, args)
}

func (r *Resolver) DeletePreciseIndexes(ctx context.Context, args *resolverstubs.DeletePreciseIndexesArgs) (*resolverstubs.EmptyResponse, error) {
	return r.autoIndexingRootResolver.DeletePreciseIndexes(ctx, args)
}

func (r *Resolver) ReindexPreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*resolverstubs.EmptyResponse, error) {
	return r.autoIndexingRootResolver.ReindexPreciseIndex(ctx, args)
}

func (r *Resolver) ReindexPreciseIndexes(ctx context.Context, args *resolverstubs.ReindexPreciseIndexesArgs) (*resolverstubs.EmptyResponse, error) {
	return r.autoIndexingRootResolver.ReindexPreciseIndexes(ctx, args)
}

func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *resolverstubs.LSIFRepositoryUploadsQueryArgs) (_ resolverstubs.LSIFUploadConnectionResolver, err error) {
	return r.uploadsRootResolver.LSIFUploadsByRepo(ctx, args)
}

func (r *Resolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	return r.uploadsRootResolver.DeleteLSIFUpload(ctx, args)
}

func (r *Resolver) DeleteLSIFUploads(ctx context.Context, args *resolverstubs.DeleteLSIFUploadsArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.uploadsRootResolver.DeleteLSIFUploads(ctx, args)
}

func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (_ resolverstubs.LSIFIndexResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexByID(ctx, id)
}

func (r *Resolver) LSIFIndexes(ctx context.Context, args *resolverstubs.LSIFIndexesQueryArgs) (_ resolverstubs.LSIFIndexConnectionResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexes(ctx, args)
}

func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *resolverstubs.LSIFRepositoryIndexesQueryArgs) (_ resolverstubs.LSIFIndexConnectionResolver, err error) {
	return r.autoIndexingRootResolver.LSIFIndexesByRepo(ctx, args)
}

func (r *Resolver) DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.DeleteLSIFIndex(ctx, args)
}

func (r *Resolver) DeleteLSIFIndexes(ctx context.Context, args *resolverstubs.DeleteLSIFIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.DeleteLSIFIndexes(ctx, args)
}

func (r *Resolver) ReindexLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.ReindexLSIFIndex(ctx, args)
}

func (r *Resolver) ReindexLSIFIndexes(ctx context.Context, args *resolverstubs.ReindexLSIFIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.ReindexLSIFIndexes(ctx, args)
}

func (r *Resolver) CommitGraph(ctx context.Context, id graphql.ID) (_ resolverstubs.CodeIntelligenceCommitGraphResolver, err error) {
	return r.uploadsRootResolver.CommitGraph(ctx, id)
}

func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.QueueAutoIndexJobsForRepoArgs) (_ []resolverstubs.LSIFIndexResolver, err error) {
	return r.autoIndexingRootResolver.QueueAutoIndexJobsForRepo(ctx, args)
}

func (r *Resolver) RequestLanguageSupport(ctx context.Context, args *resolverstubs.RequestLanguageSupportArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.RequestLanguageSupport(ctx, args)
}

func (r *Resolver) RequestedLanguageSupport(ctx context.Context) (_ []string, err error) {
	return r.autoIndexingRootResolver.RequestedLanguageSupport(ctx)
}

func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *resolverstubs.GitBlobLSIFDataArgs) (_ resolverstubs.GitBlobLSIFDataResolver, err error) {
	return r.codenavResolver.GitBlobLSIFData(ctx, args)
}

func (r *Resolver) GitBlobCodeIntelInfo(ctx context.Context, args *resolverstubs.GitTreeEntryCodeIntelInfoArgs) (_ resolverstubs.GitBlobCodeIntelSupportResolver, err error) {
	return r.autoIndexingRootResolver.GitBlobCodeIntelInfo(ctx, args)
}

func (r *Resolver) GitTreeCodeIntelInfo(ctx context.Context, args *resolverstubs.GitTreeEntryCodeIntelInfoArgs) (resolver resolverstubs.GitTreeCodeIntelSupportResolver, err error) {
	return r.autoIndexingRootResolver.GitTreeCodeIntelInfo(ctx, args)
}

func (r *Resolver) ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.policiesRootResolver.ConfigurationPolicyByID(ctx, id)
}

func (r *Resolver) CodeIntelligenceConfigurationPolicies(ctx context.Context, args *resolverstubs.CodeIntelligenceConfigurationPoliciesArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyConnectionResolver, err error) {
	return r.policiesRootResolver.CodeIntelligenceConfigurationPolicies(ctx, args)
}

func (r *Resolver) CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.CreateCodeIntelligenceConfigurationPolicyArgs) (_ resolverstubs.CodeIntelligenceConfigurationPolicyResolver, err error) {
	return r.policiesRootResolver.CreateCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *Resolver) UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.policiesRootResolver.UpdateCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *Resolver) DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *resolverstubs.DeleteCodeIntelligenceConfigurationPolicyArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.policiesRootResolver.DeleteCodeIntelligenceConfigurationPolicy(ctx, args)
}

func (r *Resolver) RepositorySummary(ctx context.Context, id graphql.ID) (_ resolverstubs.CodeIntelRepositorySummaryResolver, err error) {
	return r.autoIndexingRootResolver.RepositorySummary(ctx, id)
}

func (r *Resolver) IndexConfiguration(ctx context.Context, id graphql.ID) (_ resolverstubs.IndexConfigurationResolver, err error) {
	return r.autoIndexingRootResolver.IndexConfiguration(ctx, id)
}

func (r *Resolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *resolverstubs.UpdateRepositoryIndexConfigurationArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.UpdateRepositoryIndexConfiguration(ctx, args)
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, args *resolverstubs.PreviewRepositoryFilterArgs) (_ resolverstubs.RepositoryFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewRepositoryFilter(ctx, args)
}

func (r *Resolver) CodeIntelligenceInferenceScript(ctx context.Context) (_ string, err error) {
	return r.autoIndexingRootResolver.CodeIntelligenceInferenceScript(ctx)
}

func (r *Resolver) UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceInferenceScriptArgs) (_ *resolverstubs.EmptyResponse, err error) {
	return r.autoIndexingRootResolver.UpdateCodeIntelligenceInferenceScript(ctx, args)
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *resolverstubs.PreviewGitObjectFilterArgs) (_ resolverstubs.GitObjectFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewGitObjectFilter(ctx, id, args)
}
