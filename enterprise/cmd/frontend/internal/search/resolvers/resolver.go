package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Resolver is the GraphQL resolver of all things related to search jobs.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB) graphqlbackend.SearchJobsResolver {
	return &Resolver{logger: logger, db: db}
}

var _ graphqlbackend.SearchJobsResolver = &Resolver{}

func (r *Resolver) CreateSearchJob(ctx context.Context, args *graphqlbackend.CreateSearchJobArgs) (graphqlbackend.SearchJobResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) CancelSearchJob(ctx context.Context, args *graphqlbackend.CancelSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) DeleteSearchJob(ctx context.Context, args *graphqlbackend.DeleteSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) RetrySearchJob(ctx context.Context, args *graphqlbackend.RetrySearchJobArgs) (graphqlbackend.SearchJobResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) ValidateSearchJobQuery(ctx context.Context, args *graphqlbackend.ValidateSearchJobQueryArgs) (graphqlbackend.ValidateSearchJobQueryResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) SearchJobs(ctx context.Context, args *graphqlbackend.SearchJobsArgs) (graphqlbackend.SearchJobsConnectionResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		searchJobIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.searchJobByID(ctx, id)
		},
		searchJobRepoIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.searchJobRepoByID(ctx, id)
		},
		searchJobRepoRevisionIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.searchJobRepoRevisionByID(ctx, id)
		},
	}
}

func (r *Resolver) searchJobByID(ctx context.Context, id graphql.ID) (graphqlbackend.SearchJobResolver, error) {
	return &searchJobResolver{}, nil
}

func (r *Resolver) searchJobRepoByID(ctx context.Context, id graphql.ID) (graphqlbackend.SearchJobRepoResolver, error) {
	return &searchJobRepoResolver{}, nil
}

func (r *Resolver) searchJobRepoRevisionByID(ctx context.Context, id graphql.ID) (graphqlbackend.SearchJobRepoRevisionResolver, error) {
	return &searchJobRepoRevisionResolver{}, nil
}
