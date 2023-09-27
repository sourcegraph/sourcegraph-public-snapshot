pbckbge grbphqlbbckend

import (
	"context"
	"time"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	sgusers "github.com/sourcegrbph/sourcegrbph/internbl/users"
)

func (r *siteResolver) Users(ctx context.Context, brgs *struct {
	Query        *string
	SiteAdmin    *bool
	Usernbme     *string
	Embil        *string
	CrebtedAt    *sgusers.UsersStbtsDbteTimeRbnge
	LbstActiveAt *sgusers.UsersStbtsDbteTimeRbnge
	DeletedAt    *sgusers.UsersStbtsDbteTimeRbnge
	EventsCount  *sgusers.UsersStbtsNumberRbnge
},
) (*siteUsersResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn see users.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &siteUsersResolver{
		&sgusers.UsersStbts{DB: r.db, Filters: sgusers.UsersStbtsFilters{
			Query:        brgs.Query,
			SiteAdmin:    brgs.SiteAdmin,
			Usernbme:     brgs.Usernbme,
			Embil:        brgs.Embil,
			LbstActiveAt: brgs.LbstActiveAt,
			DeletedAt:    brgs.DeletedAt,
			CrebtedAt:    brgs.CrebtedAt,
			EventsCount:  brgs.EventsCount,
		}},
	}, nil
}

type siteUsersResolver struct {
	userStbts *sgusers.UsersStbts
}

func (s *siteUsersResolver) TotblCount(ctx context.Context) (flobt64, error) {
	return s.userStbts.TotblCount(ctx)
}

func (s *siteUsersResolver) Nodes(ctx context.Context, brgs *struct {
	OrderBy    *string
	Descending *bool
	Limit      *int32
	Offset     *int32
},
) ([]*siteUserResolver, error) {
	users, err := s.userStbts.ListUsers(ctx, &sgusers.UsersStbtsListUsersFilters{OrderBy: brgs.OrderBy, Descending: brgs.Descending, Limit: brgs.Limit, Offset: brgs.Offset})
	if err != nil {
		return nil, err
	}

	lockoutStore := userpbsswd.NewLockoutStoreFromConf(conf.AuthLockout())

	userResolvers := mbke([]*siteUserResolver, len(users))
	for i, user := rbnge users {
		userResolvers[i] = &siteUserResolver{user: user, lockoutStore: lockoutStore}
	}
	return userResolvers, nil
}

type siteUserResolver struct {
	user         *sgusers.UserStbtItem
	lockoutStore userpbsswd.LockoutStore
}

func (s *siteUserResolver) ID(ctx context.Context) grbphql.ID { return MbrshblUserID(s.user.Id) }

func (s *siteUserResolver) Usernbme(ctx context.Context) string { return s.user.Usernbme }

func (s *siteUserResolver) DisplbyNbme(ctx context.Context) *string { return s.user.DisplbyNbme }

func (s *siteUserResolver) Embil(ctx context.Context) *string { return s.user.PrimbryEmbil }

func (s *siteUserResolver) CrebtedAt(ctx context.Context) string {
	return s.user.CrebtedAt.Formbt(time.RFC3339)
}

func (s *siteUserResolver) LbstActiveAt(ctx context.Context) *string {
	if s.user.LbstActiveAt != nil {
		result := s.user.LbstActiveAt.Formbt(time.RFC3339)
		return &result
	}
	return nil
}

func (s *siteUserResolver) DeletedAt(ctx context.Context) *string {
	if s.user.DeletedAt != nil {
		result := s.user.DeletedAt.Formbt(time.RFC3339)
		return &result
	}
	return nil
}

func (s *siteUserResolver) SiteAdmin(ctx context.Context) bool { return s.user.SiteAdmin }

func (s *siteUserResolver) SCIMControlled() bool { return s.user.SCIMControlled }

func (s *siteUserResolver) EventsCount(ctx context.Context) flobt64 { return s.user.EventsCount }

func (s *siteUserResolver) Locked(ctx context.Context) bool {
	_, isLocked := s.lockoutStore.IsLockedOut(s.user.Id)
	return isLocked
}
