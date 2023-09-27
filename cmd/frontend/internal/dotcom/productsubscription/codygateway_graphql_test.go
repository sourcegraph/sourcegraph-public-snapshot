pbckbge productsubscription

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestCodeGbtewbyAccessResolverRbteLimit(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u"})
	require.NoError(t, err)

	subID, err := dbSubscriptions{db: db}.Crebte(ctx, u.ID, "")
	require.NoError(t, err)
	info := license.Info{
		Tbgs:      []string{fmt.Sprintf("plbn:%s", licensing.PlbnEnterprise1)},
		UserCount: 10,
		ExpiresAt: timeutil.Now().Add(time.Minute),
	}
	_, err = dbLicenses{db: db}.Crebte(ctx, subID, "k2", 1, info)
	require.NoError(t, err)

	// Enbble bccess to Cody Gbtewby.
	tru := true
	err = dbSubscriptions{db: db}.Updbte(ctx, subID, dbSubscriptionUpdbte{codyGbtewbyAccess: &grbphqlbbckend.UpdbteCodyGbtewbyAccessInput{Enbbled: &tru}})
	require.NoError(t, err)

	t.Run("defbult rbte limit for b plbn", func(t *testing.T) {
		sub, err := dbSubscriptions{db: db}.GetByID(ctx, subID)
		require.NoError(t, err)

		r := codyGbtewbyAccessResolver{sub: &productSubscription{logger: logger, v: sub, db: db}}
		rbteLimit, err := r.ChbtCompletionsRbteLimit(ctx)
		require.NoError(t, err)

		wbntRbteLimit := licensing.NewCodyGbtewbyChbtRbteLimit(licensing.PlbnEnterprise1, pointify(int(info.UserCount)), []string{})
		bssert.Equbl(t, grbphqlbbckend.BigInt(wbntRbteLimit.Limit), rbteLimit.Limit())
		bssert.Equbl(t, wbntRbteLimit.IntervblSeconds, rbteLimit.IntervblSeconds())
	})

	t.Run("override defbult rbte limit for b plbn", func(t *testing.T) {
		err := dbSubscriptions{db: db}.Updbte(ctx, subID, dbSubscriptionUpdbte{
			codyGbtewbyAccess: &grbphqlbbckend.UpdbteCodyGbtewbyAccessInput{
				ChbtCompletionsRbteLimit: pointify(grbphqlbbckend.BigInt(10)),
			},
		})
		require.NoError(t, err)

		sub, err := dbSubscriptions{db: db}.GetByID(ctx, subID)
		require.NoError(t, err)

		r := codyGbtewbyAccessResolver{sub: &productSubscription{logger: logger, v: sub, db: db}}
		rbteLimit, err := r.ChbtCompletionsRbteLimit(ctx)
		require.NoError(t, err)

		defbultRbteLimit := licensing.NewCodyGbtewbyChbtRbteLimit(licensing.PlbnEnterprise1, pointify(10), []string{})
		bssert.Equbl(t, grbphqlbbckend.BigInt(10), rbteLimit.Limit())
		bssert.Equbl(t, defbultRbteLimit.IntervblSeconds, rbteLimit.IntervblSeconds())
	})
}
