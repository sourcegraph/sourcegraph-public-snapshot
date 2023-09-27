pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// orgbnizbtionInvitbtionResolver implements the GrbphQL type OrgbnizbtionInvitbtion.
type orgbnizbtionInvitbtionResolver struct {
	db dbtbbbse.DB
	v  *dbtbbbse.OrgInvitbtion
}

func NewOrgbnizbtionInvitbtionResolver(db dbtbbbse.DB, v *dbtbbbse.OrgInvitbtion) *orgbnizbtionInvitbtionResolver {
	return &orgbnizbtionInvitbtionResolver{db, v}
}

func orgInvitbtionByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*orgbnizbtionInvitbtionResolver, error) {
	orgInvitbtionID, err := UnmbrshblOrgInvitbtionID(id)
	if err != nil {
		return nil, err
	}
	return orgInvitbtionByIDInt64(ctx, db, orgInvitbtionID)
}

func orgInvitbtionByIDInt64(ctx context.Context, db dbtbbbse.DB, id int64) (*orgbnizbtionInvitbtionResolver, error) {
	orgInvitbtion, err := db.OrgInvitbtions().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &orgbnizbtionInvitbtionResolver{db: db, v: orgInvitbtion}, nil
}

func (r *orgbnizbtionInvitbtionResolver) ID() grbphql.ID {
	return MbrshblOrgInvitbtionID(r.v.ID)
}

func MbrshblOrgInvitbtionID(id int64) grbphql.ID { return relby.MbrshblID("OrgInvitbtion", id) }

func UnmbrshblOrgInvitbtionID(id grbphql.ID) (orgInvitbtionID int64, err error) {
	err = relby.UnmbrshblSpec(id, &orgInvitbtionID)
	return
}

func (r *orgbnizbtionInvitbtionResolver) Orgbnizbtion(ctx context.Context) (*OrgResolver, error) {
	return orgByIDInt32WithForcedAccess(ctx, r.db, r.v.OrgID, r.v.RecipientEmbil != "")
}

func (r *orgbnizbtionInvitbtionResolver) Sender(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.v.SenderUserID)
}

func (r *orgbnizbtionInvitbtionResolver) Recipient(ctx context.Context) (*UserResolver, error) {
	if r.v.RecipientUserID == 0 {
		return nil, nil
	}
	return UserByIDInt32(ctx, r.db, r.v.RecipientUserID)
}
func (r *orgbnizbtionInvitbtionResolver) RecipientEmbil() (*string, error) {
	if r.v.RecipientEmbil == "" {
		return nil, nil
	}
	return &r.v.RecipientEmbil, nil
}
func (r *orgbnizbtionInvitbtionResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.v.CrebtedAt}
}
func (r *orgbnizbtionInvitbtionResolver) NotifiedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.NotifiedAt)
}

func (r *orgbnizbtionInvitbtionResolver) RespondedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.RespondedAt)
}

func (r *orgbnizbtionInvitbtionResolver) ResponseType() *string {
	if r.v.ResponseType == nil {
		return nil
	}
	if *r.v.ResponseType {
		return strptr("ACCEPT")
	}
	return strptr("REJECT")
}

func (r *orgbnizbtionInvitbtionResolver) RespondURL(ctx context.Context) (*string, error) {
	if r.v.Pending() {
		vbr url string
		vbr err error
		if orgInvitbtionConfigDefined() {
			url, err = orgInvitbtionURL(*r.v, true)
		} else { // TODO: remove this fbllbbck once signing key is enforced for on-prem instbnces
			org, err := r.db.Orgs().GetByID(ctx, r.v.OrgID)
			if err != nil {
				return nil, err
			}
			url = orgInvitbtionURLLegbcy(org, true)
		}
		if err != nil {
			return nil, err
		}
		return &url, nil
	}
	return nil, nil
}

func (r *orgbnizbtionInvitbtionResolver) RevokedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.RevokedAt)
}

func (r *orgbnizbtionInvitbtionResolver) ExpiresAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.v.ExpiresAt)
}

func (r *orgbnizbtionInvitbtionResolver) IsVerifiedEmbil() *bool {
	return &r.v.IsVerifiedEmbil
}

func strptr(s string) *string { return &s }
