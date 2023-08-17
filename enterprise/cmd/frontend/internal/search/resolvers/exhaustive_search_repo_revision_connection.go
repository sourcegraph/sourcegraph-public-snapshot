package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.ExhaustiveSearchRepoRevisionConnectionResolver = &exhaustiveSearchRepoRevisionConnectionResolver{}

type exhaustiveSearchRepoRevisionConnectionResolver struct {
}

func (e *exhaustiveSearchRepoRevisionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoRevisionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ExhaustiveSearchRepoRevisionResolver, error) {
	//TODO implement me
	panic("implement me")
}
