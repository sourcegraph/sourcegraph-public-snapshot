package resolvers

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func NewResolver(db database.DB) graphqlbackend.NotebooksResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db database.DB
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		"Notebook": func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.NotebookByID(ctx, id)
		},
	}
}

const notebookIDKind = "Notebook"

func marshalNotebookID(notebookID int64) graphql.ID {
	return relay.MarshalID(notebookIDKind, notebookID)
}

func unmarshalNotebookID(id graphql.ID) (notebookID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != notebookIDKind {
		err = errors.Errorf("expected graphql ID to have kind %q; got %q", notebookIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &notebookID)
	return
}

func (r *Resolver) NotebookByID(ctx context.Context, id graphql.ID) (graphqlbackend.NotebookResolver, error) {
	notebookID, err := unmarshalNotebookID(id)
	if err != nil {
		return nil, err
	}

	notebook, err := notebooks.Notebooks(r.db).GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	return &notebookResolver{notebook, r.db}, nil
}

type notebookResolver struct {
	notebook *notebooks.Notebook
	db       database.DB
}

func (r *notebookResolver) ID() graphql.ID {
	return marshalNotebookID(r.notebook.ID)
}

func (r *notebookResolver) Title(ctx context.Context) string {
	return r.notebook.Title
}

func (r *notebookResolver) Blocks(ctx context.Context) []graphqlbackend.NotebookBlockResolver {
	blockResolvers := make([]graphqlbackend.NotebookBlockResolver, 0, len(r.notebook.Blocks))
	for _, block := range r.notebook.Blocks {
		blockResolvers = append(blockResolvers, &notebookBlockResolver{block})
	}
	return blockResolvers
}

func (r *notebookResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.notebook.CreatorUserID == 0 {
		return nil, nil
	}
	return graphqlbackend.UserByIDInt32(ctx, r.db, r.notebook.CreatorUserID)
}

func (r *notebookResolver) Public(ctx context.Context) bool {
	return r.notebook.Public
}

func (r *notebookResolver) UpdatedAt(ctx context.Context) graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.notebook.UpdatedAt}
}

func (r *notebookResolver) CreatedAt(ctx context.Context) graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.notebook.CreatedAt}
}

func (r *notebookResolver) ViewerCanManage(ctx context.Context) bool {
	user, err := backend.CurrentUser(ctx, r.db)
	if err != nil {
		return false
	}
	if user == nil {
		return false
	}
	return user.ID == r.notebook.CreatorUserID
}

type notebookBlockResolver struct {
	block notebooks.NotebookBlock
}

func (r *notebookBlockResolver) ToMarkdownBlock() (graphqlbackend.MarkdownBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookMarkdownBlockType {
		return &markdownBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookBlockResolver) ToQueryBlock() (graphqlbackend.QueryBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookQueryBlockType {
		return &queryBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookBlockResolver) ToFileBlock() (graphqlbackend.FileBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookFileBlockType {
		return &fileBlockResolver{r.block}, true
	}
	return nil, false
}

type markdownBlockResolver struct {
	// block.type == NotebookMarkdownBlockType
	block notebooks.NotebookBlock
}

func (r *markdownBlockResolver) ID() string {
	return r.block.ID
}

func (r *markdownBlockResolver) MarkdownInput() string {
	return r.block.MarkdownInput.Text
}

type queryBlockResolver struct {
	// block.type == NotebookQueryBlockType
	block notebooks.NotebookBlock
}

func (r *queryBlockResolver) ID() string {
	return r.block.ID
}

func (r *queryBlockResolver) QueryInput() string {
	return r.block.QueryInput.Text
}

type fileBlockResolver struct {
	// block.type == NotebookFileBlockType
	block notebooks.NotebookBlock
}

func (r *fileBlockResolver) ID() string {
	return r.block.ID
}

func (r *fileBlockResolver) FileInput() graphqlbackend.FileBlockInputResolver {
	return &fileBlockInputResolver{*r.block.FileInput}
}

type fileBlockInputResolver struct {
	input notebooks.NotebookFileBlockInput
}

func (r *fileBlockInputResolver) RepositoryName() string {
	return r.input.RepositoryName
}

func (r *fileBlockInputResolver) FilePath() string {
	return r.input.FilePath
}

func (r *fileBlockInputResolver) Revision() *string {
	return r.input.Revision
}

func (r *fileBlockInputResolver) LineRange() graphqlbackend.FileBlockLineRangeResolver {
	if r.input.LineRange == nil {
		return nil
	}
	return &fileBlockLineRangeResolver{*r.input.LineRange}
}

type fileBlockLineRangeResolver struct {
	lineRange notebooks.LineRange
}

func (r *fileBlockLineRangeResolver) StartLine() int32 {
	return r.lineRange.StartLine
}

func (r *fileBlockLineRangeResolver) EndLine() int32 {
	return r.lineRange.EndLine
}
