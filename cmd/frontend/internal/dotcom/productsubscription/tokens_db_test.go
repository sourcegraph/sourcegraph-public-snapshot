package productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestLookupProductSubscriptionIDByAccessToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	require.NoError(t, err)

	ps, err := dbSubscriptions{db: db}.Create(ctx, u.ID, "")
	require.NoError(t, err)

	now := timeutil.Now()
	info := license.Info{
		Tags:      []string{"true-up"},
		UserCount: 10,
		ExpiresAt: now.Add(5 * time.Minute),
	}
	pl, err := dbLicenses{db: db}.Create(ctx, ps, "k", 1, info)
	require.NoError(t, err)

	t.Run("out-of-the-box token", func(t *testing.T) {
		lc, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)

		accessToken := license.GenerateLicenseKeyBasedAccessToken(lc.LicenseKey)

		gotPS, err := newDBTokens(db).LookupProductSubscriptionIDByAccessToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, gotPS, ps)
	})

	t.Run("legacy token prefix", func(t *testing.T) {
		lc, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)

		accessToken := license.GenerateLicenseKeyBasedAccessToken(lc.LicenseKey)
		accessToken = productsubscription.AccessTokenPrefix + accessToken[len(license.LicenseKeyBasedAccessTokenPrefix):]

		gotPS, err := newDBTokens(db).LookupProductSubscriptionIDByAccessToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, gotPS, ps)
	})
}
