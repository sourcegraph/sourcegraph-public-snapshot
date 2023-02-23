package graphqlbackend

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) AccessRequests(ctx context.Context, args *struct {
	database.AccessRequestsFilterArgs
	graphqlutil.ConnectionResolverArgs
}) (*graphqlutil.ConnectionResolver[*accessRequestResolver], error) {
	// 🚨 SECURITY: Only site admins can see access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	connectionStore := &accessRequestConnectionStore{
		db:   r.db,
		args: &args.AccessRequestsFilterArgs,
	}

	reverse := false
	connectionOptions := graphqlutil.ConnectionResolverOptions{
		Reverse:   &reverse,
		OrderBy:   database.OrderBy{{Field: string(database.AccessRequestListID)}},
		Ascending: false,
	}
	return graphqlutil.NewConnectionResolver[*accessRequestResolver](connectionStore, &args.ConnectionResolverArgs, &connectionOptions)
}

type accessRequestConnectionStore struct {
	db   database.DB
	args *database.AccessRequestsFilterArgs
}

func (s *accessRequestConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.AccessRequests().Count(ctx, s.args)
	if err != nil {
		return nil, err
	}

	totalCount := int32(count)

	return &totalCount, nil
}

func (s *accessRequestConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*accessRequestResolver, error) {
	accessRequests, err := s.db.AccessRequests().List(ctx, s.args, args)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*accessRequestResolver, len(accessRequests))
	for i, accessRequest := range accessRequests {
		resolvers[i] = &accessRequestResolver{accessRequest: accessRequest}
	}

	return resolvers, nil
}

func (s *accessRequestConnectionStore) MarshalCursor(node *accessRequestResolver, _ database.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New(`node is nil`)
	}

	cursor := string(node.ID())

	return &cursor, nil
}

func (s *accessRequestConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	nodeID, err := unmarshalAccessRequestID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := strconv.Itoa(int(nodeID))

	return &id, nil
}

// accessRequestResolver resolves an access request.
type accessRequestResolver struct {
	accessRequest *types.AccessRequest
}

func (s *accessRequestResolver) ID() graphql.ID { return marshalAccessRequestID(s.accessRequest.ID) }

func (s *accessRequestResolver) Name() string { return s.accessRequest.Name }

func (s *accessRequestResolver) Email() string { return s.accessRequest.Email }

func (s *accessRequestResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: s.accessRequest.CreatedAt}
}

func (s *accessRequestResolver) AdditionalInfo() *string { return &s.accessRequest.AdditionalInfo }

func (s *accessRequestResolver) Status() string { return string(s.accessRequest.Status) }

func (r *schemaResolver) SetAccessRequestStatus(ctx context.Context, args *struct {
	ID     graphql.ID
	Status types.AccessRequestStatus
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can update access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmarshalAccessRequestID(args.ID)
	if err != nil {
		return nil, err
	}

	store, err := r.db.AccessRequests().Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = store.Done(err) }()

	accessRequest, err := store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	accessRequest.Status = args.Status
	if _, err := store.Update(ctx, &types.AccessRequest{ID: accessRequest.ID, Status: accessRequest.Status}); err != nil {
		return nil, err
	}
	return nil, nil
}

func marshalAccessRequestID(id int32) graphql.ID { return relay.MarshalID("AccessRequest", id) }

func unmarshalAccessRequestID(id graphql.ID) (userID int32, err error) {
	err = relay.UnmarshalSpec(id, &userID)
	return
}
