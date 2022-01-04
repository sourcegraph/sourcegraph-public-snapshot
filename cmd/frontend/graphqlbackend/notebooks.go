package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type NotebooksOrderBy string

const (
	NotebookOrderByUpdatedAt NotebooksOrderBy = "NOTEBOOK_UPDATED_AT"
	NotebookOrderByCreatedAt NotebooksOrderBy = "NOTEBOOK_CREATED_AT"
)

type NotebooksResolver interface {
	NotebookByID(ctx context.Context, id graphql.ID) (NotebookResolver, error)
	CreateNotebook(ctx context.Context, args CreateNotebookInputArgs) (NotebookResolver, error)
	UpdateNotebook(ctx context.Context, args UpdateNotebookInputArgs) (NotebookResolver, error)
	DeleteNotebook(ctx context.Context, args DeleteNotebookArgs) (*EmptyResponse, error)
	Notebooks(ctx context.Context, args ListNotebooksArgs) (NotebookConnectionResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type NotebookConnectionResolver interface {
	Nodes(ctx context.Context) []NotebookResolver
	TotalCount(ctx context.Context) int32
	PageInfo(ctx context.Context) *graphqlutil.PageInfo
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

type NotebookBlockType string

const (
	NotebookMarkdownBlockType NotebookBlockType = "MARKDOWN"
	NotebookQueryBlockType    NotebookBlockType = "QUERY"
	NotebookFileBlockType     NotebookBlockType = "FILE"
)

type CreateNotebookInputArgs struct {
	Notebook NotebookInputArgs `json:"notebook"`
}

type UpdateNotebookInputArgs struct {
	ID       graphql.ID        `json:"id"`
	Notebook NotebookInputArgs `json:"notebook"`
}

type DeleteNotebookArgs struct {
	ID graphql.ID `json:"id"`
}

type NotebookInputArgs struct {
	Title  string                         `json:"title"`
	Blocks []CreateNotebookBlockInputArgs `json:"blocks"`
	Public bool                           `json:"public"`
}

type CreateNotebookBlockInputArgs struct {
	ID            string                `json:"id"`
	Type          NotebookBlockType     `json:"type"`
	MarkdownInput *string               `json:"markdownInput"`
	QueryInput    *string               `json:"queryInput"`
	FileInput     *CreateFileBlockInput `json:"fileInput"`
}

type CreateFileBlockInput struct {
	RepositoryName string                         `json:"repositoryName"`
	FilePath       string                         `json:"filePath"`
	Revision       *string                        `json:"revision"`
	LineRange      *CreateFileBlockLineRangeInput `json:"lineRange"`
}

type CreateFileBlockLineRangeInput struct {
	StartLine int32 `json:"startLine"`
	EndLine   int32 `json:"endLine"`
}

type ListNotebooksArgs struct {
	First         int32            `json:"first"`
	After         *string          `json:"after"`
	Query         *string          `json:"query"`
	CreatorUserID *graphql.ID      `json:"creatorUserID"`
	OrderBy       NotebooksOrderBy `json:"orderBy"`
	Descending    bool             `json:"descending"`
}
