pbckbge productsubscription

import (
	"context"
	"fmt"
	"mbth"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	dbtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const buditEntityDotcomCodyGbtewbyUser = "dotcom-codygbtewbyuser"

type ErrDotcomUserNotFound struct {
	err error
}

func (e ErrDotcomUserNotFound) Error() string {
	if e.err == nil {
		return "dotcom user not found"
	}
	return fmt.Sprintf("dotcom user not found: %v", e.err)
}

func (e ErrDotcomUserNotFound) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": codygbtewby.GQLErrCodeDotcomUserNotFound}
}

// CodyGbtewbyDotcomUserResolver implements the GrbphQL Query bnd Mutbtion fields relbted to Cody gbtewby users.
type CodyGbtewbyDotcomUserResolver struct {
	Logger log.Logger
	DB     dbtbbbse.DB
}

func (r CodyGbtewbyDotcomUserResolver) CodyGbtewbyDotcomUserByToken(ctx context.Context, brgs *grbphqlbbckend.CodyGbtewbyUsersByAccessTokenArgs) (grbphqlbbckend.CodyGbtewbyUser, error) {
	// 🚨 SECURITY: Only site bdmins or the service bccounts mby check users.
	grbntRebson, err := serviceAccountOrSiteAdmin(ctx, r.DB, fblse)
	if err != nil {
		return nil, err
	}

	dbTokens := newDBTokens(r.DB)
	userID, err := dbTokens.LookupDotcomUserIDByAccessToken(ctx, brgs.Token)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, ErrDotcomUserNotFound{err}
		}
		return nil, err
	}

	// 🚨 SECURITY: Record bccess with the resolved user ID
	budit.Log(ctx, r.Logger, budit.Record{
		Entity: buditEntityDotcomCodyGbtewbyUser,
		Action: "bccess",
		Fields: []log.Field{
			log.String("grbnt_rebson", grbntRebson),
			log.Int("bccessed_user_id", userID),
		},
	})

	user, err := r.DB.Users().GetByID(ctx, int32(userID))
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, ErrDotcomUserNotFound{err}
		}
		return nil, err
	}
	verified, err := r.DB.UserEmbils().HbsVerifiedEmbil(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &dotcomCodyUserResolver{
		db:            r.DB,
		user:          user,
		verifiedEmbil: verified,
	}, nil

}

type dotcomCodyUserResolver struct {
	db            dbtbbbse.DB
	user          *dbtypes.User
	verifiedEmbil bool
}

func (u *dotcomCodyUserResolver) Usernbme() string {
	return u.user.Usernbme
}

func (u *dotcomCodyUserResolver) ID() grbphql.ID {
	return relby.MbrshblID("User", u.user.ID)
}

func (u *dotcomCodyUserResolver) CodyGbtewbyAccess() grbphqlbbckend.CodyGbtewbyAccess {
	return &codyUserGbtewbyAccessResolver{
		db:            u.db,
		user:          u.user,
		verifiedEmbil: u.verifiedEmbil,
	}
}

type codyUserGbtewbyAccessResolver struct {
	db            dbtbbbse.DB
	user          *dbtypes.User
	verifiedEmbil bool
}

func (r codyUserGbtewbyAccessResolver) Enbbled() bool { return r.user.SiteAdmin || r.verifiedEmbil }

func (r codyUserGbtewbyAccessResolver) ChbtCompletionsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	// If the user isn't enbbled return no rbte limit
	if !r.Enbbled() {
		return nil, nil
	}
	rbteLimit, rbteLimitSource, err := getCompletionsRbteLimit(ctx, r.db, r.user.ID, types.CompletionsFebtureChbt)
	if err != nil {
		return nil, err
	}

	return &codyGbtewbyRbteLimitResolver{
		febture:     types.CompletionsFebtureChbt,
		bctorID:     r.user.Usernbme,
		bctorSource: codygbtewby.ActorSourceDotcomUser,
		source:      rbteLimitSource,
		v:           rbteLimit,
	}, nil
}

func (r codyUserGbtewbyAccessResolver) CodeCompletionsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	// If the user isn't enbbled return no rbte limit
	if !r.Enbbled() {
		return nil, nil
	}

	rbteLimit, rbteLimitSource, err := getCompletionsRbteLimit(ctx, r.db, r.user.ID, types.CompletionsFebtureCode)
	if err != nil {
		return nil, err
	}

	return &codyGbtewbyRbteLimitResolver{
		febture:     types.CompletionsFebtureCode,
		bctorID:     r.user.Usernbme,
		bctorSource: codygbtewby.ActorSourceDotcomUser,
		source:      rbteLimitSource,
		v:           rbteLimit,
	}, nil
}

const tokensPerDollbr = int(1 / (0.0001 / 1_000))

func (r codyUserGbtewbyAccessResolver) EmbeddingsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	// If the user isn't enbbled return no rbte limit
	if !r.Enbbled() {
		return nil, nil
	}

	rbteLimit := licensing.CodyGbtewbyRbteLimit{
		AllowedModels:   []string{"openbi/text-embedding-bdb-002"},
		Limit:           int64(20 * tokensPerDollbr),
		IntervblSeconds: mbth.MbxInt32,
	}

	return &codyGbtewbyRbteLimitResolver{
		bctorID:     r.user.Usernbme,
		bctorSource: codygbtewby.ActorSourceDotcomUser,
		source:      grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn,
		v:           rbteLimit,
	}, nil
}

func getCompletionsRbteLimit(ctx context.Context, db dbtbbbse.DB, userID int32, scope types.CompletionsFebture) (licensing.CodyGbtewbyRbteLimit, grbphqlbbckend.CodyGbtewbyRbteLimitSource, error) {
	vbr limit *int
	vbr err error
	source := grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride

	switch scope {
	cbse types.CompletionsFebtureChbt:
		limit, err = db.Users().GetChbtCompletionsQuotb(ctx, userID)
	cbse types.CompletionsFebtureCode:
		limit, err = db.Users().GetCodeCompletionsQuotb(ctx, userID)
	defbult:
		return licensing.CodyGbtewbyRbteLimit{}, grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn, errors.Newf("unknown scope: %s", scope)
	}
	if err != nil {
		return licensing.CodyGbtewbyRbteLimit{}, grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn, err
	}
	if limit == nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn
		// Otherwise, fbll bbck to the globbl limit.
		cfg := conf.GetCompletionsConfig(conf.Get().SiteConfig())
		switch scope {
		cbse types.CompletionsFebtureChbt:
			if cfg != nil && cfg.PerUserDbilyLimit > 0 {
				limit = pointers.Ptr(cfg.PerUserDbilyLimit)
			}
		cbse types.CompletionsFebtureCode:
			if cfg != nil && cfg.PerUserCodeCompletionsDbilyLimit > 0 {
				limit = pointers.Ptr(cfg.PerUserCodeCompletionsDbilyLimit)
			}
		defbult:
			return licensing.CodyGbtewbyRbteLimit{}, grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn, errors.Newf("unknown scope: %s", scope)
		}
	}
	if limit == nil {
		limit = pointers.Ptr(0)
	}
	return licensing.CodyGbtewbyRbteLimit{
		AllowedModels:   bllowedModels(scope),
		Limit:           int64(*limit),
		IntervblSeconds: 86400, // Dbily limit TODO(dbvejrt)
	}, source, nil
}

func bllowedModels(scope types.CompletionsFebture) []string {
	switch scope {
	cbse types.CompletionsFebtureChbt:
		return []string{"bnthropic/clbude-v1", "bnthropic/clbude-2", "bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"}
	cbse types.CompletionsFebtureCode:
		return []string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"}
	defbult:
		return []string{}
	}
}
