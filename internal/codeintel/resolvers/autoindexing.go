package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type AutoindexingServiceResolver interface {
	// Inference configuration
	CodeIntelligenceInferenceScript(ctx context.Context) (string, error)
	UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (*EmptyResponse, error)

	// Repository configuration
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*EmptyResponse, error)

	// Inference
	InferAutoIndexJobsForRepo(ctx context.Context, args *InferAutoIndexJobsForRepoArgs) ([]AutoIndexJobDescriptionResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]PreciseIndexResolver, error)

	// Coverage
	CodeIntelSummary(ctx context.Context) (CodeIntelSummaryResolver, error)
	RepositorySummary(ctx context.Context, id graphql.ID) (CodeIntelRepositorySummaryResolver, error)
	GitBlobCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (GitBlobCodeIntelSupportResolver, error)
	GitTreeCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (GitTreeCodeIntelSupportResolver, error)
}

type UpdateCodeIntelligenceInferenceScriptArgs struct {
	Script string
}

type UpdateRepositoryIndexConfigurationArgs struct {
	Repository    graphql.ID
	Configuration string
}

type InferAutoIndexJobsForRepoArgs struct {
	Repository graphql.ID
	Rev        *string
	Script     *string
}

type QueueAutoIndexJobsForRepoArgs struct {
	Repository    graphql.ID
	Rev           *string
	Configuration *string
}

type GitTreeEntryCodeIntelInfoArgs struct {
	Repo   *types.Repo
	Path   string
	Commit string
}

type IndexConfigurationResolver interface {
	Configuration(ctx context.Context) (*string, error)
	ParsedConfiguration(ctx context.Context) (*[]AutoIndexJobDescriptionResolver, error)
	InferredConfiguration(ctx context.Context) (InferredConfigurationResolver, error)
}

type InferredConfigurationResolver interface {
	Configuration() string
	ParsedConfiguration(ctx context.Context) (*[]AutoIndexJobDescriptionResolver, error)
	LimitError() *string
}

type CodeIntelSummaryResolver interface {
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int32, error)
	RepositoriesWithErrors(ctx context.Context, args *RepositoriesWithErrorsArgs) (CodeIntelRepositoryWithErrorConnectionResolver, error)
	RepositoriesWithConfiguration(ctx context.Context, args *RepositoriesWithConfigurationArgs) (CodeIntelRepositoryWithConfigurationConnectionResolver, error)
}

type (
	RepositoriesWithErrorsArgs                             = PagedConnectionArgs
	RepositoriesWithConfigurationArgs                      = PagedConnectionArgs
	CodeIntelRepositoryWithErrorConnectionResolver         = PagedConnectionWithTotalCountResolver[CodeIntelRepositoryWithErrorResolver]
	CodeIntelRepositoryWithConfigurationConnectionResolver = PagedConnectionWithTotalCountResolver[CodeIntelRepositoryWithConfigurationResolver]
)

type CodeIntelRepositoryWithErrorResolver interface {
	Repository() RepositoryResolver
	Count() int32
}

type CodeIntelRepositoryWithConfigurationResolver interface {
	Repository() RepositoryResolver
	Indexers() []IndexerWithCountResolver
}

type IndexerWithCountResolver interface {
	Indexer() CodeIntelIndexerResolver
	Count() int32
}

type CodeIntelRepositorySummaryResolver interface {
	RecentActivity(ctx context.Context) ([]PreciseIndexResolver, error)
	LastUploadRetentionScan() *gqlutil.DateTime
	LastIndexScan() *gqlutil.DateTime
	AvailableIndexers() []InferredAvailableIndexersResolver
	LimitError() *string
}

type InferredAvailableIndexersResolver interface {
	Indexer() CodeIntelIndexerResolver
	Roots() []string
	RootsWithKeys() []RootsWithKeyResolver
}

type RootsWithKeyResolver interface {
	Root() string
	ComparisonKey() string
}

type GitBlobCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (SearchBasedSupportResolver, error)
	PreciseSupport(context.Context) (PreciseSupportResolver, error)
}

type SearchBasedSupportResolver interface {
	SupportLevel() string
	Language() string
}

type PreciseSupportResolver interface {
	SupportLevel() string
	Indexers() *[]CodeIntelIndexerResolver
}

type GitTreeCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (*[]GitTreeSearchBasedCoverage, error)
	PreciseSupport(context.Context) (GitTreePreciseCoverageErrorResolver, error)
}

type GitTreeSearchBasedCoverage interface {
	CoveredPaths() []string
	Support() SearchBasedSupportResolver
}

type GitTreePreciseCoverageErrorResolver interface {
	Coverage() []GitTreePreciseCoverage
	LimitError() *string
}

type GitTreePreciseCoverage interface {
	Support() PreciseSupportResolver
	Confidence() string
}
