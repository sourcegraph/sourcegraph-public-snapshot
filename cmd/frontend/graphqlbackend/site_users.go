package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/users"
)

func (s *siteResolver) Users(ctx context.Context, args *struct {
	Query        *string
	SiteAdmin    *bool
	Username     *string
	Email        *string
	CreatedAt    *users.UsersStatsDateTimeRange
	LastActiveAt *users.UsersStatsDateTimeRange
	DeletedAt    *users.UsersStatsDateTimeRange
	EventsCount  *users.UsersStatsNumberRange
}) (*siteUsersResolver, error) {
	// 🚨 SECURITY: Only site admins can see users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, s.db); err != nil {
		return nil, err
	}

	return &siteUsersResolver{
		&users.UsersStats{DB: s.db, Filters: users.UsersStatsFilters{
			Query:        args.Query,
			SiteAdmin:    args.SiteAdmin,
			Username:     args.Username,
			Email:        args.Email,
			LastActiveAt: args.LastActiveAt,
			DeletedAt:    args.DeletedAt,
			CreatedAt:    args.CreatedAt,
			EventsCount:  args.EventsCount,
		}}}, nil
}

type siteUsersResolver struct {
	userStats *users.UsersStats
}

func (s *siteUsersResolver) TotalCount(ctx context.Context) (float64, error) {
	return s.userStats.TotalCount(ctx)
}

func (s *siteUsersResolver) Nodes(ctx context.Context, args *struct {
	OrderBy    *string
	Descending *bool
	Limit      *int32
	Offset     *int32
}) ([]*siteUserResolver, error) {
	users, err := s.userStats.ListUsers(ctx, &users.UsersStatsListUsersFilters{OrderBy: args.OrderBy, Descending: args.Descending, Limit: args.Limit, Offset: args.Offset})
	if err != nil {
		return nil, err
	}
	userResolvers := make([]*siteUserResolver, len(users))
	for i, user := range users {
		userResolvers[i] = &siteUserResolver{user}
	}
	return userResolvers, nil
}

type siteUserResolver struct {
	user *users.UserStatItem
}

func (s *siteUserResolver) ID(ctx context.Context) graphql.ID { return MarshalUserID(s.user.Id) }

func (s *siteUserResolver) Username(ctx context.Context) string { return s.user.Username }

func (s *siteUserResolver) DisplayName(ctx context.Context) *string { return s.user.DisplayName }

func (s *siteUserResolver) Email(ctx context.Context) *string { return s.user.PrimaryEmail }

func (s *siteUserResolver) CreatedAt(ctx context.Context) string {
	return s.user.CreatedAt.Format(time.RFC3339)
}

func (s *siteUserResolver) LastActiveAt(ctx context.Context) *string {
	if s.user.LastActiveAt != nil {
		result := s.user.LastActiveAt.Format(time.RFC3339)
		return &result
	}
	return nil
}

func (s *siteUserResolver) DeletedAt(ctx context.Context) *string {
	if s.user.DeletedAt != nil {
		result := s.user.DeletedAt.Format(time.RFC3339)
		return &result
	}
	return nil
}

func (s *siteUserResolver) SiteAdmin(ctx context.Context) bool { return s.user.SiteAdmin }

func (s *siteUserResolver) EventsCount(ctx context.Context) float64 { return s.user.EventsCount }
