package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type NotebooksOrderBy string

const (
	NotebookOrderByUpdatedAt NotebooksOrderBy = "NOTEBOOK_UPDATED_AT"
	NotebookOrderByCreatedAt NotebooksOrderBy = "NOTEBOOK_CREATED_AT"
	NotebookOrderByStarCount NotebooksOrderBy = "NOTEBOOK_STAR_COUNT"
)

type NotebooksResolver interface {
	NotebookByID(ctx context.Context, id graphql.ID) (NotebookResolver, error)
	CreateNotebook(ctx context.Context, args CreateNotebookInputArgs) (NotebookResolver, error)
	UpdateNotebook(ctx context.Context, args UpdateNotebookInputArgs) (NotebookResolver, error)
	DeleteNotebook(ctx context.Context, args DeleteNotebookArgs) (*EmptyResponse, error)
	Notebooks(ctx context.Context, args ListNotebooksArgs) (NotebookConnectionResolver, error)

	CreateNotebookStar(ctx context.Context, args CreateNotebookStarInputArgs) (NotebookStarResolver, error)
	DeleteNotebookStar(ctx context.Context, args DeleteNotebookStarInputArgs) (*EmptyResponse, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type NotebookConnectionResolver interface {
	Nodes(ctx context.Context) []NotebookResolver
	TotalCount(ctx context.Context) int32
	PageInfo(ctx context.Context) *gqlutil.PageInfo
}

type NotebookStarResolver interface {
	User(context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
}

type NotebookStarConnectionResolver interface {
	Nodes() []NotebookStarResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type NotebookResolver interface {
	ID() graphql.ID
	Title(ctx context.Context) string
	Blocks(ctx context.Context) []NotebookBlockResolver
	Creator(ctx context.Context) (*UserResolver, error)
	Updater(ctx context.Context) (*UserResolver, error)
	Namespace(ctx context.Context) (*NamespaceResolver, error)
	Public(ctx context.Context) bool
	UpdatedAt(ctx context.Context) gqlutil.DateTime
	CreatedAt(ctx context.Context) gqlutil.DateTime
	ViewerCanManage(ctx context.Context) (bool, error)
	ViewerHasStarred(ctx context.Context) (bool, error)
	Stars(ctx context.Context, args ListNotebookStarsArgs) (NotebookStarConnectionResolver, error)
	PatternType(ctx context.Context) string
}

type NotebookBlockResolver interface {
	ToMarkdownBlock() (MarkdownBlockResolver, bool)
	ToQueryBlock() (QueryBlockResolver, bool)
	ToFileBlock() (FileBlockResolver, bool)
	ToSymbolBlock() (SymbolBlockResolver, bool)
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

type SymbolBlockResolver interface {
	ID() string
	SymbolInput() SymbolBlockInputResolver
}

type SymbolBlockInputResolver interface {
	RepositoryName() string
	FilePath() string
	Revision() *string
	LineContext() int32
	SymbolName() string
	SymbolContainerName() string
	SymbolKind() string
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
	NotebookSymbolBlockType   NotebookBlockType = "SYMBOL"
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
	Title     string                         `json:"title"`
	Blocks    []CreateNotebookBlockInputArgs `json:"blocks"`
	Public    bool                           `json:"public"`
	Namespace graphql.ID                     `json:"namespace"`
}

type CreateNotebookBlockInputArgs struct {
	ID            string                  `json:"id"`
	Type          NotebookBlockType       `json:"type"`
	MarkdownInput *string                 `json:"markdownInput"`
	QueryInput    *string                 `json:"queryInput"`
	FileInput     *CreateFileBlockInput   `json:"fileInput"`
	SymbolInput   *CreateSymbolBlockInput `json:"symbolInput"`
}

type CreateFileBlockInput struct {
	RepositoryName string                         `json:"repositoryName"`
	FilePath       string                         `json:"filePath"`
	Revision       *string                        `json:"revision"`
	LineRange      *CreateFileBlockLineRangeInput `json:"lineRange"`
}

type CreateSymbolBlockInput struct {
	RepositoryName      string  `json:"repositoryName"`
	FilePath            string  `json:"filePath"`
	Revision            *string `json:"revision"`
	LineContext         int32   `json:"lineContext"`
	SymbolName          string  `json:"symbolName"`
	SymbolContainerName string  `json:"symbolContainerName"`
	SymbolKind          string  `json:"symbolKind"`
}

type CreateFileBlockLineRangeInput struct {
	StartLine int32 `json:"startLine"`
	EndLine   int32 `json:"endLine"`
}

type ListNotebooksArgs struct {
	First           int32            `json:"first"`
	After           *string          `json:"after"`
	Query           *string          `json:"query"`
	CreatorUserID   *graphql.ID      `json:"creatorUserID"`
	StarredByUserID *graphql.ID      `json:"starredByUserID"`
	Namespace       *graphql.ID      `json:"namespace"`
	OrderBy         NotebooksOrderBy `json:"orderBy"`
	Descending      bool             `json:"descending"`
}

type ListNotebookStarsArgs struct {
	First int32   `json:"first"`
	After *string `json:"after"`
}

type CreateNotebookStarInputArgs struct {
	NotebookID graphql.ID
}

type DeleteNotebookStarInputArgs struct {
	NotebookID graphql.ID
}
