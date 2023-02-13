package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type EmbeddingsResolver interface {
	EmbeddingsSearch(ctx context.Context, args EmbeddingsSearchInputArgs) (EmbeddingsSearchResultsResolver, error)

	ScheduleRepositoriesForEmbedding(ctx context.Context, args ScheduleRepositoriesForEmbeddingArgs) (*EmptyResponse, error)
}

type ScheduleRepositoriesForEmbeddingArgs struct {
	RepoNames []string
}

type EmbeddingsSearchInputArgs struct {
	Repository       graphql.ID
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

type EmbeddingsSearchResultsResolver interface {
	CodeResults(ctx context.Context) []EmbeddingsSearchResultResolver
	TextResults(ctx context.Context) []EmbeddingsSearchResultResolver
}

type EmbeddingsSearchResultResolver interface {
	FilePath(ctx context.Context) string
	StartLine(ctx context.Context) int32
	EndLine(ctx context.Context) int32
	Content(ctx context.Context) string
}
