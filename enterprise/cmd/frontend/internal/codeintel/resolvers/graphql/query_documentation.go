package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

func (r *QueryResolver) DocumentationPage(ctx context.Context, args *gql.LSIFDocumentationPageArgs) (gql.DocumentationPageResolver, error) {
	page, err := r.resolver.DocumentationPage(ctx, args.PageID)
	if err != nil {
		return nil, err
	}
	return NewDocumentationPageResolver(page), nil
}

type MarkupContentResolver struct {
	plaintext *string
	markdown  *string
}

func (r *MarkupContentResolver) PlainText() *string { return r.plaintext }
func (r *MarkupContentResolver) Markdown() *string  { return r.markdown }

type DocumentationNodeChildResolver struct {
	node   *DocumentationNodeResolver
	pathID string
}

func (r *DocumentationNodeChildResolver) Node() gql.DocumentationNodeChildResolver { return r.node }
func (r *DocumentationNodeChildResolver) PathID() string                           { return r.pathID }

type DocumentationNodeResolver struct {
	pathID   string
	slug     string
	newPage  bool
	tags     []string
	label    *MarkupContentResolver
	detail   *MarkupContentResolver
	children []DocumentationPageResolver
}

func (r *DocumentationNodeResolver) PathID() string                    { return r.pathID }
func (r *DocumentationNodeResolver) Slug() string                      { return r.slug }
func (r *DocumentationNodeResolver) NewPage() bool                     { return r.newPage }
func (r *DocumentationNodeResolver) Tags() []string                    { return r.tags }
func (r *DocumentationNodeResolver) Label() gql.MarkupContentResolver  { return r.label }
func (r *DocumentationNodeResolver) Detail() gql.MarkupContentResolver { return r.detail }
func (r *DocumentationNodeResolver) Children() []gql.DocumentationNodeChildResolver {
	return r.children
}

type DocumentationPageResolver struct {
	tree gql.DocumentationNodeResolver
}

func (r *DocumentationPageResolver) Tree() gql.DocumentationNodeResolver { return r.tree }

func NewDocumentationPageResolver(page *semantic.DocumentationPageData) gql.DocumentationPageResolver {
	return &DocumentationPageResolver{
		page: page,
	}
}

// TODO(slimsag): apidocs: unfuck
