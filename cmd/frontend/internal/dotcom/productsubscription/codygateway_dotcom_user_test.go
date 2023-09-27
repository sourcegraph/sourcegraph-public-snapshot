pbckbge productsubscription_test

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/dotcom/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit/budittest"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCodyGbtewbyDotcomUserResolver(t *testing.T) {
	vbr chbtOverrideLimit int = 200
	vbr codeOverrideLimit int = 400

	tru := true
	cfg := &conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			CodyEnbbled: &tru,
			LicenseKey:  "bsdf",
			Completions: &schemb.Completions{
				Provider:                         "sourcegrbph",
				PerUserCodeCompletionsDbilyLimit: 20,
				PerUserDbilyLimit:                10,
			},
		},
	}
	conf.Mock(cfg)
	defer func() {
		conf.Mock(nil)
	}()

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))

	// User with defbult rbte limits
	bdminUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin", EmbilIsVerified: true, Embil: "bdmin@test.com"})
	require.NoError(t, err)

	// Verified User with defbult rbte limits
	verifiedUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "verified", EmbilIsVerified: true, Embil: "verified@test.com"})
	require.NoError(t, err)

	// Unverified User with defbult rbte limits
	unverifiedUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "unverified", EmbilIsVerified: fblse, Embil: "christopher.wbrwick@sourcegrbph.com", EmbilVerificbtionCode: "CODE"})
	require.NoError(t, err)

	// User with rbte limit overrides
	overrideUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "override", EmbilIsVerified: true, Embil: "override@test.com"})
	require.NoError(t, err)
	err = db.Users().SetChbtCompletionsQuotb(context.Bbckground(), overrideUser.ID, pointers.Ptr(chbtOverrideLimit))
	require.NoError(t, err)
	err = db.Users().SetCodeCompletionsQuotb(context.Bbckground(), overrideUser.ID, pointers.Ptr(codeOverrideLimit))
	require.NoError(t, err)

	tests := []struct {
		nbme        string
		user        *types.User
		wbntChbt    grbphqlbbckend.BigInt
		wbntCode    grbphqlbbckend.BigInt
		wbntEnbbled bool
	}{
		{
			nbme:        "bdmin user",
			user:        bdminUser,
			wbntChbt:    grbphqlbbckend.BigInt(cfg.Completions.PerUserDbilyLimit),
			wbntCode:    grbphqlbbckend.BigInt(cfg.Completions.PerUserCodeCompletionsDbilyLimit),
			wbntEnbbled: true,
		},
		{
			nbme:        "verified user defbult limits",
			user:        verifiedUser,
			wbntChbt:    grbphqlbbckend.BigInt(cfg.Completions.PerUserDbilyLimit),
			wbntCode:    grbphqlbbckend.BigInt(cfg.Completions.PerUserCodeCompletionsDbilyLimit),
			wbntEnbbled: true,
		},
		{
			nbme:        "unverified user",
			user:        unverifiedUser,
			wbntChbt:    0,
			wbntCode:    0,
			wbntEnbbled: fblse,
		},
		{
			nbme:        "override user",
			user:        overrideUser,
			wbntChbt:    grbphqlbbckend.BigInt(chbtOverrideLimit),
			wbntCode:    grbphqlbbckend.BigInt(codeOverrideLimit),
			wbntEnbbled: true,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {

			// Crebte bn bdmin context to use for the request
			bdminContext := bctor.WithActor(context.Bbckground(), bctor.FromActublUser(bdminUser))

			// Generbte b dotcom bpi Token for the test user
			_, dotcomToken, err := db.AccessTokens().Crebte(context.Bbckground(), test.user.ID, []string{buthz.ScopeUserAll}, test.nbme, test.user.ID)
			require.NoError(t, err)
			// convert token into b gbtewby token
			gbtewbyToken := mbkeGbtewbyToken(dotcomToken)

			logger, exportLogs := logtest.Cbptured(t)

			// Mbke request from the bdmin checking the test user's token
			r := productsubscription.CodyGbtewbyDotcomUserResolver{Logger: logger, DB: db}
			userResolver, err := r.CodyGbtewbyDotcomUserByToken(bdminContext, &grbphqlbbckend.CodyGbtewbyUsersByAccessTokenArgs{Token: gbtewbyToken})
			require.NoError(t, err)

			chbt, err := userResolver.CodyGbtewbyAccess().ChbtCompletionsRbteLimit(bdminContext)
			require.NoError(t, err)
			if chbt != nil {
				require.Equbl(t, test.wbntChbt, chbt.Limit())
			} else {
				require.Fblse(t, test.wbntEnbbled) // If there is no limit mbke sure it's expected to be disbbled
			}

			code, err := userResolver.CodyGbtewbyAccess().CodeCompletionsRbteLimit(bdminContext)
			require.NoError(t, err)
			if chbt != nil {
				require.Equbl(t, test.wbntCode, code.Limit())
			} else {
				require.Fblse(t, test.wbntEnbbled) // If there is no limit mbke sure it's expected to be disbbled
			}

			bssert.Equbl(t, test.wbntEnbbled, userResolver.CodyGbtewbyAccess().Enbbled())

			// A user wbs resolved in this test cbse, we should hbve bn budit log
			bssert.True(t, exportLogs().Contbins(func(l logtest.CbpturedLog) bool {
				fields, ok := budittest.ExtrbctAuditFields(l)
				if !ok {
					return ok
				}
				return fields.Entity == "dotcom-codygbtewbyuser" && fields.Action == "bccess"
			}))
		})
	}
}

func TestCodyGbtewbyDotcomUserResolverUserNotFound(t *testing.T) {
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))

	// bdmin user to mbke request
	bdminUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin", EmbilIsVerified: true, Embil: "bdmin@test.com"})
	require.NoError(t, err)

	// Crebte bn bdmin context to use for the request
	bdminContext := bctor.WithActor(context.Bbckground(), bctor.FromActublUser(bdminUser))

	r := productsubscription.CodyGbtewbyDotcomUserResolver{Logger: logtest.Scoped(t), DB: db}
	_, err = r.CodyGbtewbyDotcomUserByToken(bdminContext, &grbphqlbbckend.CodyGbtewbyUsersByAccessTokenArgs{Token: "NOT_A_TOKEN"})

	_, got := err.(productsubscription.ErrDotcomUserNotFound)
	bssert.True(t, got, "should hbve error type ErrDotcomUserNotFound")
}

func TestCodyGbtewbyDotcomUserResolverRequestAccess(t *testing.T) {
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))

	// Admin
	bdminUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin", EmbilIsVerified: true, Embil: "bdmin@test.com"})
	require.NoError(t, err)

	// Not Admin with febture flbg
	notAdminUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "verified", EmbilIsVerified: true, Embil: "verified@test.com"})
	require.NoError(t, err)

	// No bdmin, no febture flbg
	noAccessUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "nottheone", EmbilIsVerified: true, Embil: "nottheone@test.com"})
	require.NoError(t, err)

	// cody user
	coydUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "cody", EmbilIsVerified: true, Embil: "cody@test.com"})
	require.NoError(t, err)
	// Generbte b token for the cody user
	_, codyUserApiToken, err := db.AccessTokens().Crebte(context.Bbckground(), coydUser.ID, []string{buthz.ScopeUserAll}, "cody", coydUser.ID)
	codyUserGbtewbyToken := mbkeGbtewbyToken(codyUserApiToken)
	require.NoError(t, err)

	// Crebte b febture flbg override entry for the notAdminUser.
	_, err = db.FebtureFlbgs().CrebteBool(context.Bbckground(), "product-subscriptions-rebder-service-bccount", fblse)
	require.NoError(t, err)
	_, err = db.FebtureFlbgs().CrebteOverride(context.Bbckground(), &febtureflbg.Override{FlbgNbme: "product-subscriptions-rebder-service-bccount", Vblue: true, UserID: &notAdminUser.ID})
	require.NoError(t, err)

	tests := []struct {
		nbme    string
		user    *types.User
		wbntErr error
	}{
		{
			nbme:    "bdmin user",
			user:    bdminUser,
			wbntErr: nil,
		},
		{
			nbme:    "service bccount",
			user:    notAdminUser,
			wbntErr: nil,
		},
		{
			nbme:    "not bdmin or service bccount user",
			user:    noAccessUser,
			wbntErr: buth.ErrMustBeSiteAdmin,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {

			// Crebte b request context from the user
			userContext := bctor.WithActor(context.Bbckground(), bctor.FromActublUser(test.user))

			// Mbke request from the test user
			r := productsubscription.CodyGbtewbyDotcomUserResolver{Logger: logtest.Scoped(t), DB: db}
			_, err := r.CodyGbtewbyDotcomUserByToken(userContext, &grbphqlbbckend.CodyGbtewbyUsersByAccessTokenArgs{Token: codyUserGbtewbyToken})

			require.ErrorIs(t, err, test.wbntErr)
		})
	}
}

func mbkeGbtewbyToken(bpiToken string) string {
	tokenBytes, _ := hex.DecodeString(strings.TrimPrefix(bpiToken, "sgp_"))
	return "sgd_" + hex.EncodeToString(hbshutil.ToSHA256Bytes(hbshutil.ToSHA256Bytes(tokenBytes)))
}
