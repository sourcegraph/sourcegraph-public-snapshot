package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.ExhaustiveSearchRepoConnectionResolver = &exhaustiveSearchRepoConnectionResolver{}

type exhaustiveSearchRepoConnectionResolver struct {
}

func (e *exhaustiveSearchRepoConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchRepoConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ExhaustiveSearchRepoResolver, error) {
	//TODO implement me
	panic("implement me")
}
