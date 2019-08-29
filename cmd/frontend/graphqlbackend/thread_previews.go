package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// ThreadOrThreadPreviewConnection is the interface for the GraphQL type ThreadOrThreadPreviewConnection.
type ThreadOrThreadPreviewConnection interface {
	Nodes(context.Context) ([]ToThreadOrThreadPreview, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
	Filters(context.Context) (ThreadConnectionFilters, error)
}

type ToThreadOrThreadPreview struct {
	Thread        Thread
	ThreadPreview *struct{}
	// A ThreadPreview type will be added in the future.
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

func (v ToThreadOrThreadPreview) Assignees(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ActorConnection, error) {
	return v.Common().Assignees(ctx, arg)
}

func (v ToThreadOrThreadPreview) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return v.Common().Labels(ctx, arg)
}

func (v ToThreadOrThreadPreview) ToThread() (Thread, bool) { return v.Thread, v.Thread != nil }
