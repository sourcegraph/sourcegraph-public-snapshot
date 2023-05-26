package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type AutoindexingServiceResolver interface {
	// Inference configuration
	CodeIntelligenceInferenceScript(ctx context.Context) (string, error)
	UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (*EmptyResponse, error)

	// Repository configuration
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*EmptyResponse, error)

	// Inference
	InferAutoIndexJobsForRepo(ctx context.Context, args *InferAutoIndexJobsForRepoArgs) (InferAutoIndexJobsResultResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]PreciseIndexResolver, error)
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

type InferAutoIndexJobsResultResolver interface {
	Jobs() []AutoIndexJobDescriptionResolver
	InferenceOutput() string
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
