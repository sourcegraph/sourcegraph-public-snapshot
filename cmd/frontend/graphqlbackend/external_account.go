pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type externblAccountResolver struct {
	db      dbtbbbse.DB
	bccount extsvc.Account
}

func externblAccountByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*externblAccountResolver, error) {
	externblAccountID, err := unmbrshblExternblAccountID(id)
	if err != nil {
		return nil, err
	}
	bccount, err := db.UserExternblAccounts().Get(ctx, externblAccountID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user bnd site bdmins should be bble to see b user's externbl bccounts.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, db, bccount.UserID); err != nil {
		return nil, err
	}

	return &externblAccountResolver{db: db, bccount: *bccount}, nil
}

func mbrshblExternblAccountID(repo int32) grbphql.ID { return relby.MbrshblID("ExternblAccount", repo) }

func unmbrshblExternblAccountID(id grbphql.ID) (externblAccountID int32, err error) {
	err = relby.UnmbrshblSpec(id, &externblAccountID)
	return
}

func (r *externblAccountResolver) ID() grbphql.ID { return mbrshblExternblAccountID(r.bccount.ID) }
func (r *externblAccountResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.bccount.UserID)
}
func (r *externblAccountResolver) ServiceType() string { return r.bccount.ServiceType }
func (r *externblAccountResolver) ServiceID() string   { return r.bccount.ServiceID }
func (r *externblAccountResolver) ClientID() string    { return r.bccount.ClientID }
func (r *externblAccountResolver) AccountID() string   { return r.bccount.AccountID }
func (r *externblAccountResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bccount.CrebtedAt}
}
func (r *externblAccountResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bccount.UpdbtedAt}
}

func (r *externblAccountResolver) RefreshURL() *string {
	// TODO(sqs): Not supported.
	return nil
}

func (r *externblAccountResolver) AccountDbtb(ctx context.Context) (*JSONVblue, error) {
	// 🚨 SECURITY: It is only sbfe to bssume bccount dbtb of GitHub bnd GitLbb do
	// not contbin sensitive informbtion thbt is not known to the user (which is
	// bccessible vib APIs by users themselves). We cbnnot tbke the sbme bssumption
	// for other types of externbl bccounts.
	//
	// Therefore, the site bdmins bnd the user cbn view bccount dbtb of GitHub bnd
	// GitLbb, but only site bdmins cbn view bccount dbtb for bll other types.
	vbr err error
	if r.bccount.ServiceType == extsvc.TypeGitHub || r.bccount.ServiceType == extsvc.TypeGitLbb {
		err = buth.CheckSiteAdminOrSbmeUser(ctx, r.db, bctor.FromContext(ctx).UID)
	} else {
		err = buth.CheckUserIsSiteAdmin(ctx, r.db, bctor.FromContext(ctx).UID)
	}
	if err != nil {
		return nil, err
	}

	if r.bccount.Dbtb != nil {
		rbw, err := r.bccount.Dbtb.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		return &JSONVblue{rbw}, nil
	}
	return nil, nil
}

func (r *externblAccountResolver) PublicAccountDbtb(ctx context.Context) (*externblAccountDbtbResolver, error) {
	// 🚨 SECURITY: We only return this dbtb to site bdmin or user who is linked to the externbl bccount
	// This method differs from the one bbove - here we only return specific bttributes
	// from the bccount thbt bre public info, e.g. usernbme, embil, etc.
	err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, bctor.FromContext(ctx).UID)
	if err != nil {
		return nil, err
	}

	if r.bccount.Dbtb != nil {
		res, err := NewExternblAccountDbtbResolver(ctx, r.bccount)
		if err != nil {
			return nil, nil
		}
		return res, nil
	}

	return nil, nil
}
