package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.ExhaustiveSearchesConnectionResolver = &exhaustiveSearchesConnectionResolver{}

type exhaustiveSearchesConnectionResolver struct {
}

func (e *exhaustiveSearchesConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchesConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ExhaustiveSearchResolver, error) {
	//TODO implement me
	panic("implement me")
}
