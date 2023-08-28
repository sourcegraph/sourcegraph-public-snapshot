package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.SearchJobsConnectionResolver = &searchJobsConnectionResolver{}

type searchJobsConnectionResolver struct {
}

func (e *searchJobsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 0, errors.New("not implemented")
}

func (e *searchJobsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return nil, errors.New("not implemented")
}

func (e *searchJobsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchJobResolver, error) {
	return nil, errors.New("not implemented")
}
