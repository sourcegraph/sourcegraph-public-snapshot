package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.SearchJobRepoConnectionResolver = &searchJobRepoConnectionResolver{}

type searchJobRepoConnectionResolver struct {
}

func (e *searchJobRepoConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobRepoConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchJobRepoResolver, error) {
	//TODO implement me
	panic("implement me")
}
