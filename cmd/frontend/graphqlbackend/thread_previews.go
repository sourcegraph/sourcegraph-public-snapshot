package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// ThreadPreview is the interface for the GraphQL type ThreadPreview.
type ThreadPreview interface {
	Internal_Input() CreateThreadInput
	Repository(context.Context) (*RepositoryResolver, error)
	Internal_RepositoryID() api.RepoID
	Title() string
	IsDraft() bool
	Author(context.Context) (*Actor, error)
	Body() string
	BodyText() string
	BodyHTML() string
	Kind(context.Context) (ThreadKind, error)
	RepositoryComparison(context.Context) (RepositoryComparison, error)
	Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (DiagnosticConnection, error)
	Assignable
	Labelable
	InternalID() (string, error)
}

// ThreadOrThreadPreviewConnection is the interface for the GraphQL type ThreadOrThreadPreviewConnection.
type ThreadOrThreadPreviewConnection interface {
	Nodes(context.Context) ([]ToThreadOrThreadPreview, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
	Filters(context.Context) (ThreadConnectionFilters, error)
}

type ToThreadOrThreadPreview struct {
	Thread        Thread
	ThreadPreview ThreadPreview
}

func (v ToThreadOrThreadPreview) Common() interface {
	Repository(ctx context.Context) (*RepositoryResolver, error)
	Internal_RepositoryID() api.RepoID
	Kind(context.Context) (ThreadKind, error)
	RepositoryComparison(context.Context) (RepositoryComparison, error)
	Assignable
	Labelable
} {
	switch {
	case v.Thread != nil:
		return v.Thread
	case v.ThreadPreview != nil:
		return v.ThreadPreview
	default:
		panic("invalid ToThreadOrThreadPreview")
	}
}

func (v ToThreadOrThreadPreview) Repository(ctx context.Context) (*RepositoryResolver, error) {
	return v.Common().Repository(ctx)
}

func (v ToThreadOrThreadPreview) RepositoryComparison(ctx context.Context) (RepositoryComparison, error) {
	return v.Common().RepositoryComparison(ctx)
}

func (v ToThreadOrThreadPreview) Diagnostics(ctx context.Context, args *graphqlutil.ConnectionArgs) (DiagnosticConnection, error) {
	switch {
	case v.Thread != nil:
		return v.Thread.Diagnostics(ctx, &ThreadDiagnosticConnectionArgs{ConnectionArgs: *args})
	case v.ThreadPreview != nil:
		return v.ThreadPreview.Diagnostics(ctx, args)
	default:
		panic("invalid ToThreadOrThreadPreview")
	}
}

func (v ToThreadOrThreadPreview) Assignees(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ActorConnection, error) {
	return v.Common().Assignees(ctx, arg)
}

func (v ToThreadOrThreadPreview) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return v.Common().Labels(ctx, arg)
}

func (v ToThreadOrThreadPreview) ToThread() (Thread, bool) { return v.Thread, v.Thread != nil }
func (v ToThreadOrThreadPreview) ToThreadPreview() (ThreadPreview, bool) {
	return v.ThreadPreview, v.ThreadPreview != nil
}
