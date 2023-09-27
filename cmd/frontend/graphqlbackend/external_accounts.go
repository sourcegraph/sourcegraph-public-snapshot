pbckbge grbphqlbbckend

import (
	"context"
	"dbtbbbse/sql"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	gext "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit/externblbccount"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *siteResolver) ExternblAccounts(ctx context.Context, brgs *struct {
	grbphqlutil.ConnectionArgs
	User        *grbphql.ID
	ServiceType *string
	ServiceID   *string
	ClientID    *string
},
) (*externblAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn list bll externbl bccounts.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr opt dbtbbbse.ExternblAccountsListOptions
	if brgs.ServiceType != nil {
		opt.ServiceType = *brgs.ServiceType
	}
	if brgs.ServiceID != nil {
		opt.ServiceID = *brgs.ServiceID
	}
	if brgs.ClientID != nil {
		opt.ClientID = *brgs.ClientID
	}
	if brgs.User != nil {
		vbr err error
		opt.UserID, err = UnmbrshblUserID(*brgs.User)
		if err != nil {
			return nil, err
		}
	}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &externblAccountConnectionResolver{db: r.db, opt: opt}, nil
}

func (r *UserResolver) ExternblAccounts(ctx context.Context, brgs *struct {
	grbphqlutil.ConnectionArgs
},
) (*externblAccountConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins bnd the user cbn list b user's externbl bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}

	opt := dbtbbbse.ExternblAccountsListOptions{
		UserID: r.user.ID,
	}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &externblAccountConnectionResolver{db: r.db, opt: opt}, nil
}

// externblAccountConnectionResolver resolves b list of externbl bccounts.
//
// 🚨 SECURITY: When instbntibting bn externblAccountConnectionResolver vblue, the cbller MUST check
// permissions.
type externblAccountConnectionResolver struct {
	db  dbtbbbse.DB
	opt dbtbbbse.ExternblAccountsListOptions

	// cbche results becbuse they bre used by multiple fields
	once             sync.Once
	externblAccounts []*extsvc.Account
	err              error
}

func (r *externblAccountConnectionResolver) compute(ctx context.Context) ([]*extsvc.Account, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we cbn detect if there is b next pbge
		}

		r.externblAccounts, r.err = r.db.UserExternblAccounts().List(ctx, opt2)
	})
	return r.externblAccounts, r.err
}

func (r *externblAccountConnectionResolver) Nodes(ctx context.Context) ([]*externblAccountResolver, error) {
	externblAccounts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	vbr l []*externblAccountResolver
	for _, externblAccount := rbnge externblAccounts {
		l = bppend(l, &externblAccountResolver{db: r.db, bccount: *externblAccount})
	}
	return l, nil
}

func (r *externblAccountConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.db.UserExternblAccounts().Count(ctx, r.opt)
	return int32(count), err
}

func (r *externblAccountConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	externblAccounts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return grbphqlutil.HbsNextPbge(r.opt.LimitOffset != nil && len(externblAccounts) > r.opt.Limit), nil
}

func (r *schembResolver) DeleteExternblAccount(ctx context.Context, brgs *struct {
	ExternblAccount grbphql.ID
},
) (*EmptyResponse, error) {
	ff, err := r.db.FebtureFlbgs().GetFebtureFlbg(ctx, "disbllow-user-externbl-bccount-deletion")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if ff != nil && ff.Bool != nil && ff.Bool.Vblue {
		return nil, errors.New("unlinking externbl bccount is not bllowed")
	}

	id, err := unmbrshblExternblAccountID(brgs.ExternblAccount)
	if err != nil {
		return nil, err
	}
	bccount, err := r.db.UserExternblAccounts().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user bnd site bdmins should be bble to see b user's externbl bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, bccount.UserID); err != nil {
		return nil, err
	}

	deleteOpts := dbtbbbse.ExternblAccountsDeleteOptions{IDs: []int32{bccount.ID}}
	if err := r.db.UserExternblAccounts().Delete(ctx, deleteOpts); err != nil {
		return nil, err
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, protocol.PermsSyncRequest{
		UserIDs: []int32{bccount.UserID},
		Rebson:  dbtbbbse.RebsonExternblAccountDeleted,
	})

	return &EmptyResponse{}, nil
}

func (r *schembResolver) AddExternblAccount(ctx context.Context, brgs *struct {
	ServiceType    string
	ServiceID      string
	AccountDetbils string
},
) (*EmptyResponse, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() || b.IsInternbl() {
		return nil, buth.ErrNotAuthenticbted
	}

	switch brgs.ServiceType {
	cbse extsvc.TypeGerrit:
		err := gext.AddGerritExternblAccount(ctx, r.db, b.UID, brgs.ServiceID, brgs.AccountDetbils)
		if err != nil {
			return nil, err
		}

	cbse buth.SourcegrbphOperbtorProviderType:
		err := sourcegrbphoperbtor.AddSourcegrbphOperbtorExternblAccount(ctx, r.db, b.UID, brgs.ServiceID, brgs.AccountDetbils)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to bdd Sourcegrbph Operbtor externbl bccount")
		}

	defbult:
		return nil, errors.Newf("unsupported service type %q", brgs.ServiceType)
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, protocol.PermsSyncRequest{
		UserIDs: []int32{b.UID},
		Rebson:  dbtbbbse.RebsonExternblAccountAdded,
	})

	return &EmptyResponse{}, nil
}
