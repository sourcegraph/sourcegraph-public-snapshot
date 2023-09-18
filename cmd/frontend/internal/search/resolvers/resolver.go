package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	return newSearchJobResolver(r.db, r.svc, job), nil
}

func (r *Resolver) CancelSearchJob(ctx context.Context, args *graphqlbackend.CancelSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	jobID, err := UnmarshalSearchJobID(args.ID)
	if err != nil {
		return nil, err
	}

	return nil, r.svc.CancelSearchJob(ctx, jobID)
}

func (r *Resolver) DeleteSearchJob(ctx context.Context, args *graphqlbackend.DeleteSearchJobArgs) (*graphqlbackend.EmptyResponse, error) {
	jobID, err := UnmarshalSearchJobID(args.ID)
	if err != nil {
		return nil, err
	}

	return nil, r.svc.DeleteSearchJob(ctx, jobID)
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

	query := ""
	if args.Query != nil {
		query = *args.Query
	}

	s := &searchJobsConnectionStore{
		db:      db,
		service: service,
		states:  states,
		query:   query,
		userIDs: ids,
	}
	return graphqlutil.NewConnectionResolver[graphqlbackend.SearchJobResolver](
		s,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			Ascending: !args.Descending,
			OrderBy:   database.OrderBy{{Field: strings.ToLower(args.OrderBy)}, {Field: "id"}}},
	)
}

type searchJobsConnectionStore struct {
	db      database.DB
	service *service.Service
	states  []string
	query   string
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
		resolvers = append(resolvers, newSearchJobResolver(s.db, s.service, job))
	}

	return resolvers, nil
}

const searchJobsCursorKind = "SearchJobsCursor"

func (s *searchJobsConnectionStore) MarshalCursor(node graphqlbackend.SearchJobResolver, orderBy database.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}

	column := orderBy[0].Field

	var value string
	switch column {
	case "created_at":
		value = fmt.Sprintf("'%v'", node.CreatedAt().Format(time.RFC3339))
	case "state":
		value = fmt.Sprintf("'%v'", strings.ToLower(node.State()))
	case "query":
		value = fmt.Sprintf("'%v'", node.Query())
	default:
		return nil, errors.New(fmt.Sprintf("invalid OrderBy.Field. Expected one of (created_at, state, query). Actual: %s", column))
	}

	id, err := UnmarshalSearchJobID(node.ID())
	if err != nil {
		return nil, err
	}

	cursor := string(relay.MarshalID(
		searchJobsCursorKind,
		&types.Cursor{Column: column,
			Value: fmt.Sprintf("%s@%d", value, id)},
	))
	return &cursor, nil
}

func (s *searchJobsConnectionStore) UnmarshalCursor(cursor string, orderBy database.OrderBy) (*string, error) {
	if kind := relay.UnmarshalKind(graphql.ID(cursor)); kind != searchJobsCursorKind {
		return nil, errors.New(fmt.Sprintf("expected a %q cursor, got %q", searchJobsCursorKind, kind))
	}
	var spec *types.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(cursor), &spec); err != nil {
		return nil, err
	}

	if len(orderBy) == 0 {
		return nil, errors.New("no OrderBy provided")
	}
	column := orderBy[0].Field
	if spec.Column != column {
		return nil, errors.New(fmt.Sprintf("expected a %q cursor, got %q", column, spec.Column))
	}

	i := strings.LastIndex(spec.Value, "@")
	if i == -1 {
		return nil, errors.New(fmt.Sprintf("Invalid cursor. Expected Value: <%s>@<id> Actual Value: %s", column, spec.Value))
	}

	values := []string{spec.Value[0:i], spec.Value[i+1:]}

	csv := ""
	switch column {
	case "created_at":
		csv = fmt.Sprintf("%v, %v", values[0], values[1])
	case "state":
		csv = fmt.Sprintf("%v, %v", values[0], values[1])
	case "query":
		csv = fmt.Sprintf("%v, %v", values[0], values[1])
	default:
		return nil, errors.New("Invalid OrderBy Field.")
	}

	return &csv, nil
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
	return newSearchJobResolver(r.db, r.svc, job), nil
}
