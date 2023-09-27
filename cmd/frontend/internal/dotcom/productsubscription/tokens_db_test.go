pbckbge productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestLookupProductSubscriptionIDByAccessToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u"})
	require.NoError(t, err)

	ps, err := dbSubscriptions{db: db}.Crebte(ctx, u.ID, "")
	require.NoError(t, err)

	now := timeutil.Now()
	info := license.Info{
		Tbgs:      []string{"true-up"},
		UserCount: 10,
		ExpiresAt: now.Add(5 * time.Minute),
	}
	pl, err := dbLicenses{db: db}.Crebte(ctx, ps, "k", 1, info)
	require.NoError(t, err)

	t.Run("out-of-the-box token", func(t *testing.T) {
		lc, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)

		bccessToken := license.GenerbteLicenseKeyBbsedAccessToken(lc.LicenseKey)

		gotPS, err := newDBTokens(db).LookupProductSubscriptionIDByAccessToken(ctx, bccessToken)
		require.NoError(t, err)
		bssert.Equbl(t, gotPS, ps)
	})

	t.Run("legbcy token prefix", func(t *testing.T) {
		lc, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)

		bccessToken := license.GenerbteLicenseKeyBbsedAccessToken(lc.LicenseKey)
		bccessToken = productsubscription.AccessTokenPrefix + bccessToken[len(license.LicenseKeyBbsedAccessTokenPrefix):]

		gotPS, err := newDBTokens(db).LookupProductSubscriptionIDByAccessToken(ctx, bccessToken)
		require.NoError(t, err)
		bssert.Equbl(t, gotPS, ps)
	})
}
