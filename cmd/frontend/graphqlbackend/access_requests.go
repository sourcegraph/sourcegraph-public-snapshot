package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// TODO: add tests
func MarshalAccessRequestID(id int32) graphql.ID { return relay.MarshalID("AccessRequest", id) }

func UnmarshalAccessRequestID(id graphql.ID) (userID int32, err error) {
	err = relay.UnmarshalSpec(id, &userID)
	return
}

func (r *schemaResolver) SetAccessRequestStatus(ctx context.Context, args *struct {
	ID     graphql.ID
	Status types.AccessRequestStatus
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can update access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := UnmarshalAccessRequestID(args.ID)
	if err != nil {
		return nil, err
	}
	accessRequest, err := r.db.AccessRequests().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	accessRequest.Status = args.Status
	if err := r.db.AccessRequests().Update(ctx, database.AccessRequestUpdate{ID: accessRequest.ID, Status: &accessRequest.Status}); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *schemaResolver) AccessRequests(ctx context.Context, args *database.AccessRequestsFilterOptions) (*accessRequestsResolver, error) {
	// 🚨 SECURITY: Only site admins can see access requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &accessRequestsResolver{r.db, args}, nil
}

type accessRequestsResolver struct {
	db            database.DB
	filterOptions *database.AccessRequestsFilterOptions
}

func (s *accessRequestsResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := s.db.AccessRequests().Count(ctx, *s.filterOptions)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (s *accessRequestsResolver) Nodes(ctx context.Context, args *database.AccessRequestsListOptions) ([]*accessRequestResolver, error) {
	accessRequests, err := s.db.AccessRequests().List(ctx, database.AccessRequestsFilterAndListOptions{
		AccessRequestsFilterOptions: *s.filterOptions,
		AccessRequestsListOptions:   *args,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*accessRequestResolver, len(accessRequests))
	for i, accessRequest := range accessRequests {
		resolvers[i] = &accessRequestResolver{accessRequest: accessRequest}
	}

	return resolvers, nil
}

type accessRequestResolver struct {
	accessRequest *types.AccessRequest
}

func (s *accessRequestResolver) ID() graphql.ID { return MarshalAccessRequestID(s.accessRequest.ID) }

func (s *accessRequestResolver) Name(ctx context.Context) string { return s.accessRequest.Name }

func (s *accessRequestResolver) Email(ctx context.Context) string { return s.accessRequest.Email }

func (s *accessRequestResolver) CreatedAt(ctx context.Context) gqlutil.DateTime {
	return gqlutil.DateTime{Time: s.accessRequest.CreatedAt}
}

func (s *accessRequestResolver) AdditionalInfo(ctx context.Context) *string {
	return &s.accessRequest.AdditionalInfo
}
func (s *accessRequestResolver) Status(ctx context.Context) string {
	return string(s.accessRequest.Status)
}
