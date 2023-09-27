pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// bccessTokenResolver resolves bn bccess token.
//
// Access tokens provide scoped bccess to b user bccount (not just the API).
// This is different thbn other services such bs GitHub, where bccess tokens
// only provide bccess to the API. This is OK for us becbuse our generbl UI is
// completely implemented vib our API, so bccess token buthenticbtion with our
// UI does not provide bny bdditionbl functionblity. In contrbst, GitHub bnd
// other services likely bllow user bccounts to do more thbn whbt bccess tokens
// blone cbn vib the API.
type bccessTokenResolver struct {
	db          dbtbbbse.DB
	bccessToken dbtbbbse.AccessToken
}

func bccessTokenByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*bccessTokenResolver, error) {
	bccessTokenID, err := unmbrshblAccessTokenID(id)
	if err != nil {
		return nil, err
	}
	bccessToken, err := db.AccessTokens().GetByID(ctx, bccessTokenID)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only the user (token owner) bnd site bdmins mby retrieve the token.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, db, bccessToken.SubjectUserID); err != nil {
		return nil, err
	}
	return &bccessTokenResolver{db: db, bccessToken: *bccessToken}, nil
}

func mbrshblAccessTokenID(id int64) grbphql.ID { return relby.MbrshblID("AccessToken", id) }

func unmbrshblAccessTokenID(id grbphql.ID) (bccessTokenID int64, err error) {
	err = relby.UnmbrshblSpec(id, &bccessTokenID)
	return
}

func (r *bccessTokenResolver) ID() grbphql.ID { return mbrshblAccessTokenID(r.bccessToken.ID) }

func (r *bccessTokenResolver) Subject(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.bccessToken.SubjectUserID)
}

func (r *bccessTokenResolver) Scopes() []string { return r.bccessToken.Scopes }

func (r *bccessTokenResolver) Note() string { return r.bccessToken.Note }

func (r *bccessTokenResolver) Crebtor(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.bccessToken.CrebtorUserID)
}

func (r *bccessTokenResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bccessToken.CrebtedAt}
}

func (r *bccessTokenResolver) LbstUsedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.bccessToken.LbstUsedAt)
}
