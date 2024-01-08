package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AccessRequestsArgs struct {
	database.AccessRequestsFilterArgs
	graphqlutil.ConnectionResolverArgs
}

func (r *schemaResolver) AccessRequests(ctx context.Context, args *AccessRequestsArgs) (*graphqlutil.ConnectionResolver[*accessRequestResolver], error) {
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

func (s *accessRequestConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.AccessRequests().Count(ctx, s.args)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
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

func (s *accessRequestConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	nodeID, err := unmarshalAccessRequestID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{nodeID}, nil
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

	err = r.db.WithTransact(ctx, func(tx database.DB) error {
		store := tx.AccessRequests()

		accessRequest, err := store.GetByID(ctx, id)
		if err != nil {
			return err
		}

		currentUser, err := auth.CurrentUser(ctx, tx)
		if err != nil {
			return err
		}

		accessRequest.Status = args.Status
		if _, err := store.Update(ctx, &types.AccessRequest{ID: accessRequest.ID, Status: accessRequest.Status, DecisionByUserID: &currentUser.ID}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func accessRequestByID(ctx context.Context, db database.DB, id graphql.ID) (*accessRequestResolver, error) {
	// 🚨 SECURITY: Only site admins can see access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	accessRequestID, err := unmarshalAccessRequestID(id)
	if err != nil {
		return nil, err
	}
	accessRequest, err := db.AccessRequests().GetByID(ctx, accessRequestID)
	if err != nil {
		return nil, err
	}

	return &accessRequestResolver{accessRequest}, nil
}

func marshalAccessRequestID(id int32) graphql.ID { return relay.MarshalID("AccessRequest", id) }

func unmarshalAccessRequestID(id graphql.ID) (userID int32, err error) {
	err = relay.UnmarshalSpec(id, &userID)
	return
}
