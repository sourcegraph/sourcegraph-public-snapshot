package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	sgusers "github.com/sourcegraph/sourcegraph/internal/users"
)

func (r *siteResolver) Users(ctx context.Context, args *struct {
	Query        *string
	SiteAdmin    *bool
	Username     *string
	Email        *string
	CreatedAt    *sgusers.UsersStatsDateTimeRange
	LastActiveAt *sgusers.UsersStatsDateTimeRange
	DeletedAt    *sgusers.UsersStatsDateTimeRange
	EventsCount  *sgusers.UsersStatsNumberRange
},
) (*siteUsersResolver, error) {
	// 🚨 SECURITY: Only site admins can see users.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &siteUsersResolver{
		&sgusers.UsersStats{DB: r.db, Filters: sgusers.UsersStatsFilters{
			Query:        args.Query,
			SiteAdmin:    args.SiteAdmin,
			Username:     args.Username,
			Email:        args.Email,
			LastActiveAt: args.LastActiveAt,
			DeletedAt:    args.DeletedAt,
			CreatedAt:    args.CreatedAt,
			EventsCount:  args.EventsCount,
		}},
	}, nil
}

type siteUsersResolver struct {
	userStats *sgusers.UsersStats
}

func (s *siteUsersResolver) TotalCount(ctx context.Context) (float64, error) {
	return s.userStats.TotalCount(ctx)
}

func (s *siteUsersResolver) Nodes(ctx context.Context, args *struct {
	OrderBy    *string
	Descending *bool
	Limit      *int32
	Offset     *int32
},
) ([]*siteUserResolver, error) {
	users, err := s.userStats.ListUsers(ctx, &sgusers.UsersStatsListUsersFilters{OrderBy: args.OrderBy, Descending: args.Descending, Limit: args.Limit, Offset: args.Offset})
	if err != nil {
		return nil, err
	}

	lockoutStore := userpasswd.NewLockoutStoreFromConf(conf.AuthLockout())

	userResolvers := make([]*siteUserResolver, len(users))
	for i, user := range users {
		userResolvers[i] = &siteUserResolver{user: user, lockoutStore: lockoutStore}
	}
	return userResolvers, nil
}

type siteUserResolver struct {
	user         *sgusers.UserStatItem
	lockoutStore userpasswd.LockoutStore
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

func (s *siteUserResolver) SCIMControlled() bool { return s.user.SCIMControlled }

func (s *siteUserResolver) EventsCount(ctx context.Context) float64 { return s.user.EventsCount }

func (s *siteUserResolver) Locked(ctx context.Context) bool {
	_, isLocked := s.lockoutStore.IsLockedOut(s.user.Id)
	return isLocked
}
