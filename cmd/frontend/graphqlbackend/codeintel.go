package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	resolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

type CodeIntelResolver interface {
	GitBlobLSIFData(ctx context.Context, args *codenavgraphql.GitBlobLSIFDataArgs) (codenavgraphql.GitBlobLSIFDataResolver, error)
	GitBlobCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (_ autoindexinggraphql.GitBlobCodeIntelSupportResolver, err error)
	GitTreeCodeIntelInfo(ctx context.Context, args *autoindexinggraphql.GitTreeEntryCodeIntelInfoArgs) (resolver autoindexinggraphql.GitTreeCodeIntelSupportResolver, err error)
	RequestLanguageSupport(ctx context.Context, args *autoindexinggraphql.RequestLanguageSupportArgs) (*resolvers.EmptyResponse, error)
	RequestedLanguageSupport(ctx context.Context) ([]string, error)

	NodeResolvers() map[string]NodeByIDFunc

	AutoindexingServiceResolver
	ExecutorResolver
	UploadsServiceResolver
	PoliciesServiceResolver
}

type ExecutorResolver interface {
	ExecutorResolver() executor.Resolver
}

type AutoindexingServiceResolver interface {
	IndexConfiguration(ctx context.Context, id graphql.ID) (autoindexinggraphql.IndexConfigurationResolver, error) // TODO - rename ...ForRepo
	DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*resolvers.EmptyResponse, error)
	LSIFIndexByID(ctx context.Context, id graphql.ID) (_ resolvers.LSIFIndexResolver, err error)
	LSIFIndexes(ctx context.Context, args *autoindexinggraphql.LSIFIndexesQueryArgs) (resolvers.LSIFIndexConnectionResolver, error)
	LSIFIndexesByRepo(ctx context.Context, args *autoindexinggraphql.LSIFRepositoryIndexesQueryArgs) (resolvers.LSIFIndexConnectionResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *autoindexinggraphql.QueueAutoIndexJobsForRepoArgs) ([]resolvers.LSIFIndexResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *autoindexinggraphql.UpdateRepositoryIndexConfigurationArgs) (*resolvers.EmptyResponse, error)
	RepositorySummary(ctx context.Context, id graphql.ID) (resolvers.CodeIntelRepositorySummaryResolver, error)
}

type UploadsServiceResolver interface {
	CommitGraph(ctx context.Context, id graphql.ID) (uploadsgraphql.CodeIntelligenceCommitGraphResolver, error)
	LSIFUploadByID(ctx context.Context, id graphql.ID) (resolvers.LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *uploadsgraphql.LSIFUploadsQueryArgs) (resolvers.LSIFUploadConnectionResolver, error)
	LSIFUploadsByRepo(ctx context.Context, args *uploadsgraphql.LSIFRepositoryUploadsQueryArgs) (resolvers.LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*resolvers.EmptyResponse, error)
	DeleteLSIFUploads(ctx context.Context, args *uploadsgraphql.DeleteLSIFUploadsArgs) (*resolvers.EmptyResponse, error)
}
type PoliciesServiceResolver interface {
	CodeIntelligenceConfigurationPolicies(ctx context.Context, args *policiesgraphql.CodeIntelligenceConfigurationPoliciesArgs) (policiesgraphql.CodeIntelligenceConfigurationPolicyConnectionResolver, error)
	ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, error)
	CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.CreateCodeIntelligenceConfigurationPolicyArgs) (policiesgraphql.CodeIntelligenceConfigurationPolicyResolver, error)
	DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.DeleteCodeIntelligenceConfigurationPolicyArgs) (*sharedresolvers.EmptyResponse, error)
	PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *policiesgraphql.PreviewGitObjectFilterArgs) ([]policiesgraphql.GitObjectFilterPreviewResolver, error)
	PreviewRepositoryFilter(ctx context.Context, args *policiesgraphql.PreviewRepositoryFilterArgs) (policiesgraphql.RepositoryFilterPreviewResolver, error)
	UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *policiesgraphql.UpdateCodeIntelligenceConfigurationPolicyArgs) (*sharedresolvers.EmptyResponse, error)
}
