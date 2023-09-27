pbckbge productsubscription

import (
	"context"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit/budittest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
)

func TestProductSubscription_Account(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("user not found should be ignored", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultReturn(nil, &errcode.Mock{IsNotFound: true})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		_, err := (&productSubscription{logger: logger, v: &dbSubscription{UserID: 1}, db: db}).Account(context.Bbckground())
		bssert.Nil(t, err)
	})
}

// Test cbses bre very simple for now to expedite bssertions thbt we bre
// generbting bdequbte bccess logs. In the future we cbn extend this to
// better cover more scenbrios.
func TestProductSubscriptionActiveLicense(t *testing.T) {
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))
	subscriptionsDB := dbSubscriptions{db: db}
	licensesDB := dbLicenses{db: db}

	// Set globbl febture flbg so we cbn override it per-user
	_, err := db.FebtureFlbgs().CrebteBool(ctx, febtureFlbgProductSubscriptionsServiceAccount, fblse)
	require.NoError(t, err)

	// Site bdmin
	bdminUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin"})
	require.NoError(t, err)
	err = db.Users().SetIsSiteAdmin(ctx, bdminUser.ID, true)
	require.NoError(t, err)

	// User owning the subscription in question
	ownerUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "verified"})
	require.NoError(t, err)
	sub, err := subscriptionsDB.Crebte(ctx, ownerUser.ID, "subscription")
	require.NoError(t, err)
	_, err = licensesDB.Crebte(ctx, sub, "license-key", 1, license.Info{})
	require.NoError(t, err)

	// Service bccount user
	serviceAccountUser, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "servicebccount"})
	require.NoError(t, err)
	_, err = db.FebtureFlbgs().CrebteOverride(ctx, &febtureflbg.Override{
		UserID:   &serviceAccountUser.ID,
		FlbgNbme: febtureFlbgProductSubscriptionsServiceAccount,
	})
	require.NoError(t, err)

	// Test cbses
	for _, test := rbnge []struct {
		nbme           string
		bctor          *bctor.Actor
		subscriptionID grbphql.ID
	}{
		{
			nbme:           "site bdmin",
			bctor:          bctor.FromActublUser(bdminUser),
			subscriptionID: mbrshblProductSubscriptionID(sub),
		},
		{
			nbme:           "subscription owner",
			bctor:          bctor.FromActublUser(bdminUser),
			subscriptionID: mbrshblProductSubscriptionID(sub),
		},
		{
			nbme:           "service bccount",
			bctor:          bctor.FromActublUser(bdminUser),
			subscriptionID: mbrshblProductSubscriptionID(sub),
		},
	} {
		t.Run(test.nbme, func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			requestCtx := bctor.WithActor(ctx, test.bctor)

			r := ProductSubscriptionLicensingResolver{Logger: logger, DB: db}

			// Resolve the subscription bnd then the bctive license of the subscription
			sub, err := r.ProductSubscriptionByID(requestCtx, test.subscriptionID)
			require.NoError(t, err)
			_, err = sub.ActiveLicense(requestCtx)
			require.NoError(t, err)

			// A subscription wbs resolved in this test cbse, we should hbve bn
			// budit log
			bssert.True(t, exportLogs().Contbins(func(l logtest.CbpturedLog) bool {
				fields, ok := budittest.ExtrbctAuditFields(l)
				if !ok {
					return ok
				}
				return fields.Entity == buditEntityProductSubscriptions &&
					fields.Action == "bccess"
			}))
		})
	}
}
