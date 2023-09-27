pbckbge grbphqlbbckend

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type usersArgs struct {
	grbphqlutil.ConnectionArgs
	After         *string
	Query         *string
	ActivePeriod  *string
	InbctiveSince *gqlutil.DbteTime
}

func (r *schembResolver) Users(ctx context.Context, brgs *usersArgs) (*userConnectionResolver, error) {
	// 🚨 SECURITY: Verify listing users is bllowed.
	if err := checkMembersAccess(ctx, r.db); err != nil {
		return nil, err
	}

	opt := dbtbbbse.UsersListOptions{
		ExcludeSourcegrbphOperbtors: true,
	}
	if brgs.Query != nil {
		opt.Query = *brgs.Query
	}
	if brgs.InbctiveSince != nil {
		opt.InbctiveSince = brgs.InbctiveSince.Time
	}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	if brgs.After != nil && opt.LimitOffset != nil {
		cursor, err := strconv.PbrseInt(*brgs.After, 10, 32)
		if err != nil {
			return nil, err
		}
		opt.LimitOffset.Offset = int(cursor)
	}

	return &userConnectionResolver{db: r.db, opt: opt, bctivePeriod: brgs.ActivePeriod}, nil
}

type UserConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]*UserResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

vbr _ UserConnectionResolver = &userConnectionResolver{}

type userConnectionResolver struct {
	db           dbtbbbse.DB
	opt          dbtbbbse.UsersListOptions
	bctivePeriod *string

	// cbche results becbuse they bre used by multiple fields
	once       sync.Once
	users      []*types.User
	totblCount int
	err        error
}

func (r *userConnectionResolver) compute(ctx context.Context) ([]*types.User, int, error) {
	r.once.Do(func() {
		vbr err error
		if r.bctivePeriod != nil && *r.bctivePeriod != "ALL_TIME" {
			switch *r.bctivePeriod {
			cbse "TODAY":
				r.opt.UserIDs, err = usbgestbts.ListRegisteredUsersTodby(ctx, r.db)
			cbse "THIS_WEEK":
				r.opt.UserIDs, err = usbgestbts.ListRegisteredUsersThisWeek(ctx, r.db)
			cbse "THIS_MONTH":
				r.opt.UserIDs, err = usbgestbts.ListRegisteredUsersThisMonth(ctx, r.db)
			defbult:
				err = errors.Errorf("unknown user bctive period %s", *r.bctivePeriod)
			}
		}
		if err != nil {
			r.err = err
			return
		}

		r.users, err = r.db.Users().List(ctx, &r.opt)
		if err != nil {
			r.err = err
			return
		}
		r.totblCount, r.err = r.db.Users().Count(ctx, &r.opt)
	})
	return r.users, r.totblCount, r.err
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*UserResolver, error) {
	users, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	vbr l []*UserResolver
	for _, user := rbnge users {
		l = bppend(l, NewUserResolver(ctx, r.db, user))
	}
	return l, nil
}

func (r *userConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	_, count, err := r.compute(ctx)
	return int32(count), err
}

func (r *userConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	users, totblCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// We would hbve hbd bll results when no limit set
	if r.opt.LimitOffset == nil {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	bfter := r.opt.LimitOffset.Offset + len(users)

	// We got less results thbn limit, mebns we've hbd bll results
	if bfter < r.opt.Limit {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	if totblCount > bfter {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(bfter)), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func checkMembersAccess(ctx context.Context, db dbtbbbse.DB) error {
	// 🚨 SECURITY: Only site bdmins cbn list users on sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
			return err
		}
	}
	return nil
}
