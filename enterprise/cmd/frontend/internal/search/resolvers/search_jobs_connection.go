package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.SearchJobsConnectionResolver = &searchJobsConnectionResolver{}

type searchJobsConnectionResolver struct {
}

func (e *searchJobsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *searchJobsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchJobResolver, error) {
	//TODO implement me
	panic("implement me")
}
