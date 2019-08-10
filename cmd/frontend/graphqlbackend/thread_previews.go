package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// ThreadPreview is the interface for the GraphQL type ThreadPreview.
type ThreadPreview interface {
	Repository(context.Context) (*RepositoryResolver, error)
	Title() string
	Author(context.Context) (*Actor, error)
	Body() string
	BodyText() string
	BodyHTML() string
	Kind(context.Context) (ThreadKind, error)
	RepositoryComparison(context.Context) (*RepositoryComparisonResolver, error)
	Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (DiagnosticConnection, error)
}

// ThreadOrThreadPreviewConnection is the interface for the GraphQL type ThreadOrThreadPreviewConnection.
type ThreadOrThreadPreviewConnection interface {
	Nodes(context.Context) ([]ToThreadOrThreadPreview, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type ToThreadOrThreadPreview struct {
	Thread        Thread
	ThreadPreview ThreadPreview
}

func (v ToThreadOrThreadPreview) Repository(ctx context.Context) (*RepositoryResolver, error) {
	switch {
	case v.Thread != nil:
		return v.Thread.Repository(ctx)
	case v.ThreadPreview != nil:
		return v.ThreadPreview.Repository(ctx)
	default:
		panic("invalid ToThreadOrThreadPreview")
	}
}

func (v ToThreadOrThreadPreview) ToThread() (Thread, bool) { return v.Thread, v.Thread != nil }
func (v ToThreadOrThreadPreview) ToThreadPreview() (ThreadPreview, bool) {
	return v.ThreadPreview, v.ThreadPreview != nil
}
