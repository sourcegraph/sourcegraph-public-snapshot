package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.SearchJobRepoRevisionConnectionResolver = &searchJobRepoRevisionConnectionResolver{}

type searchJobRepoRevisionConnectionResolver struct {
}

func (e *searchJobRepoRevisionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoRevisionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchJobRepoRevisionResolver, error) {
	//TODO implement me
	panic("implement me")
}
