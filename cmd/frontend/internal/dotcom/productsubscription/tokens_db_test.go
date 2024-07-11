package productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
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

		gotPS, err := NewTokensDB(db).LookupProductSubscriptionIDByAccessToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, gotPS, ps)
	})

	t.Run("legacy token prefix", func(t *testing.T) {
		lc, err := dbLicenses{db: db}.GetByID(ctx, pl)
		require.NoError(t, err)

		accessToken := license.GenerateLicenseKeyBasedAccessToken(lc.LicenseKey)
		accessToken = productsubscription.AccessTokenPrefix + accessToken[len(license.LicenseKeyBasedAccessTokenPrefix):]

		gotPS, err := NewTokensDB(db).LookupProductSubscriptionIDByAccessToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, gotPS, ps)
	})

	t.Run("last_used_at Updates", func(t *testing.T) {
		// Create a new access token.
		subject, err := db.Users().Create(ctx, database.NewUser{
			Email:                 "u1@example.com",
			Username:              "u1",
			Password:              "p1",
			EmailVerificationCode: "c1",
		})
		if err != nil {
			t.Fatal(err)
		}
		creator, err := db.Users().Create(ctx, database.NewUser{
			Email:                 "u2@example.com",
			Username:              "u2",
			Password:              "p2",
			EmailVerificationCode: "c2",
		})
		if err != nil {
			t.Fatal(err)
		}

		testTokenID, testTokenValue, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b", "c"}, "n0", creator.ID, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		// Fetches the test access token. Confirm its default state has last_used_at of nil.
		initialToken, err := db.AccessTokens().GetByID(ctx, testTokenID)
		if err != nil {
			t.Fatal(err)
		}
		if initialToken.LastUsedAt != nil {
			t.Fatal("last_used_at was not nil upon token creation")
		}

		dbTokens := NewTokensDB(db)

		// Call LookupDotcomUserIDByAccessToken. This will have a side-effect of updating the
		// token's last_used_at column.
		token, err := accesstoken.GenerateDotcomUserGatewayAccessToken(testTokenValue)
		if err != nil {
			t.Fatalf("Generating dotcom user gateway token: %v", err)
		}
		gotUserID, err := dbTokens.LookupDotcomUserIDByAccessToken(ctx, token)
		if err != nil {
			t.Fatalf("Looking up dotcom User by Access Token: %v", err)
		}
		if gotUserID != int(subject.ID) {
			t.Errorf("LookupDotcomUserIDByAccessToken returned unexpected user ID: %d", gotUserID)
		}

		// Now lookup the token and confirm that last_used_at was updated as expected.
		currentToken, err := db.AccessTokens().GetByID(ctx, testTokenID)
		if err != nil {
			t.Fatal(err)
		}
		if currentToken.LastUsedAt == nil {
			t.Fatal("last_used_at was not set after calling LookupDotcomUserIDByAccessToken")
		}
		if time.Since(*currentToken.LastUsedAt) > 2*time.Second {
			t.Errorf("last_used_at was updated, but it seems to have the wrong timestamp.")
		}

		// Cleanup
		err = db.AccessTokens().DeleteByID(ctx, testTokenID)
		if err != nil {
			t.Fatal(err)
		}
	})
}
