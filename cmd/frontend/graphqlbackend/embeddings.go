package graphqlbackend

import (
	"context"
)

type EmbeddingsResolver interface {
	EmbeddingsSearch(ctx context.Context, args EmbeddingsSearchInputArgs) (EmbeddingsSearchResultsResolver, error)
	IsContextRequiredForQuery(ctx context.Context, args IsContextRequiredForQueryInputArgs) (bool, error)

	ScheduleRepositoriesForEmbedding(ctx context.Context, args ScheduleRepositoriesForEmbeddingArgs) (*EmptyResponse, error)
	ScheduleContextDetectionForEmbedding(ctx context.Context) (*EmptyResponse, error)
}

type ScheduleRepositoriesForEmbeddingArgs struct {
	RepoNames []string
}

type IsContextRequiredForQueryInputArgs struct {
	RepoName string
	Query    string
}

type EmbeddingsSearchInputArgs struct {
	RepoName         string
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

type EmbeddingsSearchResultsResolver interface {
	CodeResults(ctx context.Context) []EmbeddingsSearchResultResolver
	TextResults(ctx context.Context) []EmbeddingsSearchResultResolver
}

type EmbeddingsSearchResultResolver interface {
	FileName(ctx context.Context) string
	StartLine(ctx context.Context) int32
	EndLine(ctx context.Context) int32
	Content(ctx context.Context) string
}
