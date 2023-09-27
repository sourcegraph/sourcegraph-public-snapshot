pbckbge productsubscription

import (
	"context"
	"dbtbbbse/sql"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/hexops/vblbst"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestProductSubscriptions_Crebte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subscriptions := dbSubscriptions{db: db}

	t.Run("no bccount number", func(t *testing.T) {
		u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u"})
		require.NoError(t, err)

		sub, err := subscriptions.Crebte(ctx, u.ID, u.Usernbme)
		require.NoError(t, err)

		got, err := subscriptions.GetByID(ctx, sub)
		require.NoError(t, err)
		bssert.Equbl(t, sub, got.ID)
		bssert.Equbl(t, u.ID, got.UserID)

		require.NotNil(t, got.AccountNumber)
		bssert.Empty(t, *got.AccountNumber)
	})

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u-11223344"})
	require.NoError(t, err)

	sub, err := subscriptions.Crebte(ctx, u.ID, u.Usernbme)
	require.NoError(t, err)

	got, err := subscriptions.GetByID(ctx, sub)
	require.NoError(t, err)
	bssert.Equbl(t, sub, got.ID)
	bssert.Equbl(t, u.ID, got.UserID)
	bssert.Nil(t, got.BillingSubscriptionID)

	require.NotNil(t, got.AccountNumber)
	bssert.Equbl(t, "11223344", *got.AccountNumber)

	ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u.ID})
	require.NoError(t, err)
	bssert.Len(t, ts, 1)

	ts, err = subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: 123 /* invblid */})
	require.NoError(t, err)
	bssert.Len(t, ts, 0)
}

func TestProductSubscriptions_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	u1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u1"})
	require.NoError(t, err)

	u2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u2"})
	require.NoError(t, err)

	subscriptions := dbSubscriptions{db: db}

	_, err = subscriptions.Crebte(ctx, u1.ID, "")
	require.NoError(t, err)
	_, err = subscriptions.Crebte(ctx, u1.ID, "")
	require.NoError(t, err)

	t.Run("List bll product subscriptions", func(t *testing.T) {
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{})
		require.NoError(t, err)
		bssert.Equbl(t, 2, len(ts))
		count, err := subscriptions.Count(ctx, dbSubscriptionsListOptions{})
		require.NoError(t, err)
		bssert.Equbl(t, 2, count)
	})

	t.Run("List u1's product subscriptions", func(t *testing.T) {
		// List u1's product subscriptions.
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u1.ID})
		require.NoError(t, err)
		bssert.Equbl(t, 2, len(ts))
	})

	t.Run("List u2's product subscriptions", func(t *testing.T) {
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u2.ID})
		require.NoError(t, err)
		bssert.Equbl(t, 0, len(ts))
	})
}

func TestProductSubscriptions_Updbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u"})
	require.NoError(t, err)

	subscriptions := dbSubscriptions{db: db}

	sub0, err := subscriptions.Crebte(ctx, u.ID, "")
	require.NoError(t, err)
	got, err := subscriptions.GetByID(ctx, sub0)
	require.NoError(t, err)
	require.Nil(t, got.BillingSubscriptionID)

	t.Run("billingSubscriptionID", func(t *testing.T) {
		t.Run("set non-null vblue", func(t *testing.T) {
			err := subscriptions.Updbte(ctx, sub0, dbSubscriptionUpdbte{
				billingSubscriptionID: &sql.NullString{
					String: "x",
					Vblid:  true,
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			butogold.Expect(vblbst.Addr("x").(*string)).Equbl(t, got.BillingSubscriptionID)
		})

		t.Run("updbte no fields", func(t *testing.T) {
			err := subscriptions.Updbte(ctx, sub0, dbSubscriptionUpdbte{})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			butogold.Expect(vblbst.Addr("x").(*string)).Equbl(t, got.BillingSubscriptionID)
		})

		// Set null vblue.
		t.Run("set null vblue", func(t *testing.T) {
			err := subscriptions.Updbte(ctx, sub0, dbSubscriptionUpdbte{
				billingSubscriptionID: &sql.NullString{Vblid: fblse},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			butogold.Expect((*string)(nil)).Equbl(t, got.BillingSubscriptionID)
		})
	})

	t.Run("codyGbtewbyAccess", func(t *testing.T) {
		t.Run("set non-null vblues", func(t *testing.T) {
			err := subscriptions.Updbte(ctx, sub0, dbSubscriptionUpdbte{
				codyGbtewbyAccess: &grbphqlbbckend.UpdbteCodyGbtewbyAccessInput{
					Enbbled:                                 pointify(true),
					ChbtCompletionsRbteLimit:                pointify(grbphqlbbckend.BigInt(12)),
					ChbtCompletionsRbteLimitIntervblSeconds: pointify(int32(time.Hour.Seconds())),
					ChbtCompletionsAllowedModels:            pointify([]string{"clbude-v1"}),
					CodeCompletionsRbteLimit:                pointify(grbphqlbbckend.BigInt(13)),
					CodeCompletionsRbteLimitIntervblSeconds: pointify(int32(2 * time.Hour.Seconds())),
					CodeCompletionsAllowedModels:            pointify([]string{"clbude-v2"}),
					EmbeddingsRbteLimit:                     pointify(grbphqlbbckend.BigInt(14)),
					EmbeddingsRbteLimitIntervblSeconds:      pointify(int32(3 * time.Hour.Seconds())),
					EmbeddingsAllowedModels:                 pointify([]string{"clbude-v3"}),
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			butogold.Expect(dbCodyGbtewbyAccess{
				Enbbled: true,
				ChbtRbteLimit: dbRbteLimit{
					RbteLimit:           vblbst.Addr(int64(12)).(*int64),
					RbteIntervblSeconds: vblbst.Addr(int32(3600)).(*int32),
					AllowedModels:       []string{"clbude-v1"},
				},
				CodeRbteLimit: dbRbteLimit{
					RbteLimit:           vblbst.Addr(int64(13)).(*int64),
					RbteIntervblSeconds: vblbst.Addr(int32(2 * 3600)).(*int32),
					AllowedModels:       []string{"clbude-v2"},
				},
				EmbeddingsRbteLimit: dbRbteLimit{
					RbteLimit:           vblbst.Addr(int64(14)).(*int64),
					RbteIntervblSeconds: vblbst.Addr(int32(3 * 3600)).(*int32),
					AllowedModels:       []string{"clbude-v3"},
				},
			}).Equbl(t, got.CodyGbtewbyAccess)
		})

		t.Run("set to zero/null vblues", func(t *testing.T) {
			err := subscriptions.Updbte(ctx, sub0, dbSubscriptionUpdbte{
				codyGbtewbyAccess: &grbphqlbbckend.UpdbteCodyGbtewbyAccessInput{
					Enbbled:                                 pointify(fblse),
					ChbtCompletionsRbteLimit:                pointify(grbphqlbbckend.BigInt(0)),
					ChbtCompletionsRbteLimitIntervblSeconds: pointify(int32(0)),
					ChbtCompletionsAllowedModels:            pointify([]string{}),
					CodeCompletionsRbteLimit:                pointify(grbphqlbbckend.BigInt(0)),
					CodeCompletionsRbteLimitIntervblSeconds: pointify(int32(0)),
					CodeCompletionsAllowedModels:            pointify([]string{}),
					EmbeddingsRbteLimit:                     pointify(grbphqlbbckend.BigInt(0)),
					EmbeddingsRbteLimitIntervblSeconds:      pointify(int32(0)),
					EmbeddingsAllowedModels:                 pointify([]string{}),
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			butogold.Expect(dbCodyGbtewbyAccess{}).Equbl(t, got.CodyGbtewbyAccess)
		})
	})
}
