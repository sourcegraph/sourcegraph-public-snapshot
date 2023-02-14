package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func (r *schemaResolver) AccessRequests(ctx context.Context, args *database.AccessRequestsFilterOptions) (*accessRequestsResolver, error) {
	// 🚨 SECURITY: Only site admins can see users.
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

func (s *accessRequestResolver) ID() graphql.ID { return MarshalUserID(s.accessRequest.ID) }

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
