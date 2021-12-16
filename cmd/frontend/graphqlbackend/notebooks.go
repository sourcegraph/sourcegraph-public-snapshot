package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type NotebooksResolver interface {
	NotebookByID(ctx context.Context, id graphql.ID) (NotebookResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type NotebookResolver interface {
	ID() graphql.ID
	Title(ctx context.Context) string
	Blocks(ctx context.Context) []NotebookBlockResolver
	Creator(ctx context.Context) (*UserResolver, error)
	Public(ctx context.Context) bool
	UpdatedAt(ctx context.Context) DateTime
	CreatedAt(ctx context.Context) DateTime
	ViewerCanManage(ctx context.Context) bool
}

type NotebookBlockResolver interface {
	ToMarkdownBlock() (MarkdownBlockResolver, bool)
	ToQueryBlock() (QueryBlockResolver, bool)
	ToFileBlock() (FileBlockResolver, bool)
}

type MarkdownBlockResolver interface {
	ID() string
	MarkdownInput() string
}

type QueryBlockResolver interface {
	ID() string
	QueryInput() string
}

type FileBlockResolver interface {
	ID() string
	FileInput() FileBlockInputResolver
}

type FileBlockInputResolver interface {
	RepositoryName() string
	FilePath() string
	Revision() *string
	LineRange() FileBlockLineRangeResolver
}

type FileBlockLineRangeResolver interface {
	StartLine() int32
	EndLine() int32
}
