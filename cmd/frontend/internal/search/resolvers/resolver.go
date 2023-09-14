package resolvers

import (
	"context"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
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

func newSearchJobConnectionResolver(db database.DB, service *service.Service, args *graphqlbackend.SearchJobsArgs) (*graphqlutil.ConnectionResolver[graphqlbackend.SearchJobResolver], error) {
	var states []string
	if args.States != nil {
		states = *args.States
	}

	var ids []int32
	if args.UserIDs != nil {
		for _, id := range *args.UserIDs {
			userID, err := graphqlbackend.UnmarshalUserID(id)
			if err != nil {
				return nil, err
			}
			ids = append(ids, userID)
		}
	}

	s := &searchJobsConnectionStore{
		db:      db,
		service: service,
		states:  states,
		query:   args.Query,
		userIDs: ids,
	}
	return graphqlutil.NewConnectionResolver[graphqlbackend.SearchJobResolver](
		s,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			Ascending: !args.Descending,
			OrderBy:   database.OrderBy{{Field: strings.ToLower(args.OrderBy)}}},
	)
}

type searchJobsConnectionStore struct {
	ctx     context.Context
	db      database.DB
	service *service.Service
	states  []string
	query   *string
	userIDs []int32
}

func (s *searchJobsConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	// TODO (stefan) add "Count" method to service
	jobs, err := s.service.ListSearchJobs(ctx, store.ListArgs{States: s.states, Query: s.query, UserIDs: s.userIDs})
	if err != nil {
		return nil, err
	}

	total := int32(len(jobs))
	return &total, nil
}

func (s *searchJobsConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.SearchJobResolver, error) {
	jobs, err := s.service.ListSearchJobs(ctx, store.ListArgs{PaginationArgs: args, States: s.states, Query: s.query, UserIDs: s.userIDs})
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.SearchJobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, &searchJobResolver{
			db:  s.db,
			Job: job,
		})
	}

	return resolvers, nil
}

func (s *searchJobsConnectionStore) MarshalCursor(node graphqlbackend.SearchJobResolver, _ database.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}
	cursor := string(node.ID())
	return &cursor, nil
}

func (s *searchJobsConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	nodeID, err := UnmarshalSearchJobID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}
	id := strconv.Itoa(int(nodeID))
	return &id, nil
}

func (r *Resolver) SearchJobs(_ context.Context, args *graphqlbackend.SearchJobsArgs) (*graphqlutil.ConnectionResolver[graphqlbackend.SearchJobResolver], error) {
	return newSearchJobConnectionResolver(r.db, r.svc, args)
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
