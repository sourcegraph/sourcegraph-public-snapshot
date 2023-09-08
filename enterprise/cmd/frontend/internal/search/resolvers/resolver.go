package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to search jobs.
type Resolver struct {
	logger log.Logger
	db     database.DB
	svc    *service.Service
}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB, svc *service.Service) graphqlbackend.SearchJobsResolver {
	return &Resolver{logger: logger, db: db, svc: svc}
}

var _ graphqlbackend.SearchJobsResolver = &Resolver{}

func (r *Resolver) CreateSearchJob(ctx context.Context, args *graphqlbackend.CreateSearchJobArgs) (graphqlbackend.SearchJobResolver, error) {
	job, err := r.svc.CreateSearchJob(ctx, args.Query)
	if err != nil {
		return nil, err
	}
	return searchJobResolver{Job: job, db: r.db}, nil
}

func (r *Resolver) CancelSearchJob(ctx context.Context, args *graphqlbackend.CancelSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	jobID, err := UnmarshalSearchJobID(args.ID)
	if err != nil {
		return nil, err
	}

	return nil, r.svc.CancelSearchJob(ctx, jobID)
}

func (r *Resolver) DeleteSearchJob(ctx context.Context, args *graphqlbackend.DeleteSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	return nil, errors.New("not implemented")
}

func (r *Resolver) SearchJobs(ctx context.Context, args *graphqlbackend.SearchJobsArgs) (graphqlbackend.SearchJobsConnectionResolver, error) {
	// TODO respect args. For now we always return everything a user created
	jobs, err := r.svc.ListSearchJobs(ctx)
	if err != nil {
		return nil, err
	}
	return &searchJobsConnectionResolver{Jobs: jobs, db: r.db}, nil
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		searchJobIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.searchJobByID(ctx, id)
		},
	}
}

func (r *Resolver) searchJobByID(ctx context.Context, id graphql.ID) (graphqlbackend.SearchJobResolver, error) {
	jobID, err := UnmarshalSearchJobID(id)
	if err != nil {
		return nil, err
	}
	job, err := r.svc.GetSearchJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return searchJobResolver{Job: job, db: r.db}, nil
}
