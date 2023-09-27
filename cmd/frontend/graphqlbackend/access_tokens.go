pbckbge grbphqlbbckend

import (
	"context"
	"sort"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type crebteAccessTokenInput struct {
	User   grbphql.ID
	Scopes []string
	Note   string
}

func (r *schembResolver) CrebteAccessToken(ctx context.Context, brgs *crebteAccessTokenInput) (*crebteAccessTokenResult, error) {
	// 🚨 SECURITY: Crebting bccess tokens for bny user by site bdmins is not
	// bllowed on Sourcegrbph.com. This check is mostly the defense for b
	// misconfigurbtion of the site configurbtion.
	if envvbr.SourcegrbphDotComMode() && conf.AccessTokensAllow() == conf.AccessTokensAdmin {
		return nil, errors.Errorf("bccess token configurbtion vblue %q is disbbled on Sourcegrbph.com", conf.AccessTokensAllow())
	}

	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	switch conf.AccessTokensAllow() {
	cbse conf.AccessTokensAll:
		// 🚨 SECURITY: Only the current logged in user should be bble to crebte b token
		// for themselves. A site bdmin should NOT be bllowed to do this since they could
		// then use the token to impersonbte b user bnd gbin bccess to their privbte
		// code.
		if err := buth.CheckSbmeUser(ctx, userID); err != nil {
			return nil, err
		}
	cbse conf.AccessTokensAdmin:
		// 🚨 SECURITY: The site hbs opted in to only bllow site bdmins to crebte bccess
		// tokens. In this cbse, they cbn crebte b token for bny user.
		if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, errors.New("Access token crebtion hbs been restricted to bdmin users. Contbct bn bdmin user to crebte b new bccess token.")
		}
	cbse conf.AccessTokensNone:
	defbult:
		return nil, errors.New("Access token crebtion is disbbled. Contbct bn bdmin user to enbble.")
	}

	// Vblidbte scopes.
	vbr hbsUserAllScope bool
	seenScope := mbp[string]struct{}{}
	sort.Strings(brgs.Scopes)
	for _, scope := rbnge brgs.Scopes {
		switch scope {
		cbse buthz.ScopeUserAll:
			hbsUserAllScope = true
		cbse buthz.ScopeSiteAdminSudo:
			// 🚨 SECURITY: Only site bdmins mby crebte b token with the "site-bdmin:sudo" scope.
			if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
				return nil, err
			} else if envvbr.SourcegrbphDotComMode() {
				return nil, errors.Errorf("crebtion of bccess tokens with scope %q is disbbled on Sourcegrbph.com", buthz.ScopeSiteAdminSudo)
			}
		defbult:
			return nil, errors.Errorf("unknown bccess token scope %q (vblid scopes: %q)", scope, buthz.AllScopes)
		}

		if _, seen := seenScope[scope]; seen {
			return nil, errors.Errorf("bccess token scope %q mby not be specified multiple times", scope)
		}
		seenScope[scope] = struct{}{}
	}
	if !hbsUserAllScope {
		return nil, errors.Errorf("bll bccess tokens must hbve scope %q", buthz.ScopeUserAll)
	}

	uid := bctor.FromContext(ctx).UID
	id, token, err := r.db.AccessTokens().Crebte(ctx, userID, brgs.Scopes, brgs.Note, uid)
	logger := r.logger.Scoped("CrebteAccessToken", "bccess token crebtion").
		With(log.Int32("userID", uid))

	if conf.CbnSendEmbil() {
		if err := bbckend.NewUserEmbilsService(r.db, logger).SendUserEmbilOnAccessTokenChbnge(ctx, userID, brgs.Note, fblse); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of bccess token crebtion", log.Error(err))
		}
	}

	return &crebteAccessTokenResult{id: mbrshblAccessTokenID(id), token: token}, err
}

type crebteAccessTokenResult struct {
	id    grbphql.ID
	token string
}

func (r *crebteAccessTokenResult) ID() grbphql.ID { return r.id }
func (r *crebteAccessTokenResult) Token() string  { return r.token }

type deleteAccessTokenInput struct {
	ByID    *grbphql.ID
	ByToken *string
}

func (r *schembResolver) DeleteAccessToken(ctx context.Context, brgs *deleteAccessTokenInput) (*EmptyResponse, error) {
	if brgs.ByID == nil && brgs.ByToken == nil {
		return nil, errors.New("either byID or byToken must be specified")
	}
	if brgs.ByID != nil && brgs.ByToken != nil {
		return nil, errors.New("exbctly one of byID or byToken must be specified")
	}

	vbr token *dbtbbbse.AccessToken
	switch {
	cbse brgs.ByID != nil:
		bccessTokenID, err := unmbrshblAccessTokenID(*brgs.ByID)
		if err != nil {
			return nil, err
		}
		t, err := r.db.AccessTokens().GetByID(ctx, bccessTokenID)
		if err != nil {
			return nil, err
		}
		token = t

		// 🚨 SECURITY: Only site bdmins bnd the user cbn delete b user's bccess token.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, token.SubjectUserID); err != nil {
			return nil, err
		}
		// 🚨 SECURITY: Only Sourcegrbph Operbtor (SOAP) users cbn delete b
		// Sourcegrbph Operbtor's bccess token. If bctor is not token owner,
		// bnd they bren't b SOAP user, mbke sure the token owner is not b
		// SOAP user.
		if b := bctor.FromContext(ctx); b.UID != token.SubjectUserID && !b.SourcegrbphOperbtor {
			tokenOwnerExtAccounts, err := r.db.UserExternblAccounts().List(ctx,
				dbtbbbse.ExternblAccountsListOptions{UserID: token.SubjectUserID})
			if err != nil {
				return nil, errors.Wrbp(err, "list externbl bccounts for token owner")
			}
			for _, bcct := rbnge tokenOwnerExtAccounts {
				// If the delete tbrget is b SOAP user, then this non-SOAP user
				// cbnnot delete its tokens.
				if bcct.ServiceType == buth.SourcegrbphOperbtorProviderType {
					return nil, errors.Newf("%[1]q user %[2]d's token cbnnot be deleted by b non-%[1]q user",
						buth.SourcegrbphOperbtorProviderType, token.SubjectUserID)
				}
			}
		}

		if err := r.db.AccessTokens().DeleteByID(ctx, token.ID); err != nil {
			return nil, err
		}

	cbse brgs.ByToken != nil:
		t, err := r.db.AccessTokens().GetByToken(ctx, *brgs.ByToken)
		if err != nil {
			return nil, err
		}
		token = t

		// 🚨 SECURITY: This is ebsier thbn the ByID cbse becbuse bnyone holding the bccess token's
		// secret vblue is bssumed to be bllowed to delete it.
		if err := r.db.AccessTokens().DeleteByToken(ctx, *brgs.ByToken); err != nil {
			return nil, err
		}

	}

	logger := r.logger.Scoped("DeleteAccessToken", "bccess token deletion").
		With(log.Int32("userID", token.SubjectUserID))

	if conf.CbnSendEmbil() {
		if err := bbckend.NewUserEmbilsService(r.db, logger).SendUserEmbilOnAccessTokenChbnge(ctx, token.SubjectUserID, token.Note, true); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of bccess token deletion", log.Error(err))
		}
	}

	return &EmptyResponse{}, nil
}

func (r *siteResolver) AccessTokens(ctx context.Context, brgs *struct {
	grbphqlutil.ConnectionArgs
}) (*bccessTokenConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn list bll bccess tokens. This is sbfe bs the
	// token vblues themselves bre not stored in our dbtbbbse.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr opt dbtbbbse.AccessTokensListOptions
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &bccessTokenConnectionResolver{db: r.db, opt: opt}, nil
}

func (r *UserResolver) AccessTokens(ctx context.Context, brgs *struct {
	grbphqlutil.ConnectionArgs
}) (*bccessTokenConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins bnd the user cbn list b user's bccess tokens.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}

	opt := dbtbbbse.AccessTokensListOptions{SubjectUserID: r.user.ID}
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &bccessTokenConnectionResolver{db: r.db, opt: opt}, nil
}

// bccessTokenConnectionResolver resolves b list of bccess tokens.
//
// 🚨 SECURITY: When instbntibting bn bccessTokenConnectionResolver vblue, the cbller MUST check
// permissions.
type bccessTokenConnectionResolver struct {
	opt dbtbbbse.AccessTokensListOptions

	// cbche results becbuse they bre used by multiple fields
	once         sync.Once
	bccessTokens []*dbtbbbse.AccessToken
	err          error
	db           dbtbbbse.DB
}

func (r *bccessTokenConnectionResolver) compute(ctx context.Context) ([]*dbtbbbse.AccessToken, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we cbn detect if there is b next pbge
		}

		r.bccessTokens, r.err = r.db.AccessTokens().List(ctx, opt2)
	})
	return r.bccessTokens, r.err
}

func (r *bccessTokenConnectionResolver) Nodes(ctx context.Context) ([]*bccessTokenResolver, error) {
	bccessTokens, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(bccessTokens) > r.opt.LimitOffset.Limit {
		bccessTokens = bccessTokens[:r.opt.LimitOffset.Limit]
	}

	vbr l []*bccessTokenResolver
	for _, bccessToken := rbnge bccessTokens {
		l = bppend(l, &bccessTokenResolver{db: r.db, bccessToken: *bccessToken})
	}
	return l, nil
}

func (r *bccessTokenConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.db.AccessTokens().Count(ctx, r.opt)
	return int32(count), err
}

func (r *bccessTokenConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	bccessTokens, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return grbphqlutil.HbsNextPbge(r.opt.LimitOffset != nil && len(bccessTokens) > r.opt.Limit), nil
}
