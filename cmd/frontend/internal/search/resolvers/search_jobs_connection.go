package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
)

var _ graphqlbackend.SearchJobsConnectionResolver = &searchJobsConnectionResolver{}

type searchJobsConnectionResolver struct {
	Jobs []*types.ExhaustiveSearchJob
	db   database.DB
}

func (e *searchJobsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(e.Jobs)), nil
}

func (e *searchJobsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (e *searchJobsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchJobResolver, error) {
	resolvers := make([]graphqlbackend.SearchJobResolver, 0, len(e.Jobs))
	for _, job := range e.Jobs {
		resolvers = append(resolvers, searchJobResolver{Job: job, db: e.db})
	}
	return resolvers, nil
}
