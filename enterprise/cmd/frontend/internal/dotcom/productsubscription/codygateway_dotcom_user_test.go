package productsubscription_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCodyGatewayDotcomUserResolver(t *testing.T) {
	var chatOverrideLimit int = 200
	var codeOverrideLimit int = 400

	cfg := &conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Completions: &schema.Completions{
				PerUserCodeCompletionsDailyLimit: 20,
				PerUserDailyLimit:                10,
			},
		},
	}
	conf.Mock(cfg)
	defer func() {
		conf.Mock(nil)
	}()

	ctx := context.Background()
	db := database.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))

	// User with default rate limits
	adminUser, err := db.Users().Create(ctx, database.NewUser{Username: "admin", EmailIsVerified: true, Email: "admin@test.com"})
	require.NoError(t, err)

	// Verified User with default rate limits
	verifiedUser, err := db.Users().Create(ctx, database.NewUser{Username: "verified", EmailIsVerified: true, Email: "verified@test.com"})
	require.NoError(t, err)

	// Unverified User with default rate limits
	unverifiedUser, err := db.Users().Create(ctx, database.NewUser{Username: "unverified", EmailIsVerified: false, Email: "christopher.warwick@sourcegraph.com", EmailVerificationCode: "CODE"})
	require.NoError(t, err)

	// User with rate limit overrides
	overrideUser, err := db.Users().Create(ctx, database.NewUser{Username: "override", EmailIsVerified: true, Email: "override@test.com"})
	require.NoError(t, err)
	err = db.Users().SetChatCompletionsQuota(context.Background(), overrideUser.ID, iPtr(chatOverrideLimit))
	require.NoError(t, err)
	err = db.Users().SetCodeCompletionsQuota(context.Background(), overrideUser.ID, iPtr(codeOverrideLimit))
	require.NoError(t, err)

	tests := []struct {
		name        string
		user        *types.User
		wantChat    int32
		wantCode    int32
		wantEnabled bool
	}{
		{
			name:        "admin user",
			user:        adminUser,
			wantChat:    int32(cfg.Completions.PerUserDailyLimit),
			wantCode:    int32(cfg.Completions.PerUserCodeCompletionsDailyLimit),
			wantEnabled: true,
		},
		{
			name:        "verified user default limits",
			user:        verifiedUser,
			wantChat:    int32(cfg.Completions.PerUserDailyLimit),
			wantCode:    int32(cfg.Completions.PerUserCodeCompletionsDailyLimit),
			wantEnabled: true,
		},
		{
			name:        "unverified user",
			user:        unverifiedUser,
			wantChat:    0,
			wantCode:    0,
			wantEnabled: false,
		},
		{
			name:        "override user",
			user:        overrideUser,
			wantChat:    int32(chatOverrideLimit),
			wantCode:    int32(codeOverrideLimit),
			wantEnabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Create an admin context to use for the request
			adminContext := actor.WithActor(context.Background(), actor.FromActualUser(adminUser))

			// Generate a token for the test user
			_, token, err := db.AccessTokens().Create(context.Background(), test.user.ID, []string{authz.ScopeUserAll}, test.name, test.user.ID)
			require.NoError(t, err)

			// Make request from the admin checking the test user's token
			r := productsubscription.CodyGatewayDotcomUserResolver{DB: db}
			userResolver, err := r.DotcomCodyGatewayUserByToken(adminContext, &graphqlbackend.CodyGatewayUsersByAccessTokenArgs{Token: token})
			require.NoError(t, err)

			chat, err := userResolver.CodyGatewayAccess().ChatCompletionsRateLimit(adminContext)
			require.NoError(t, err)
			if chat != nil {
				require.Equal(t, test.wantChat, chat.Limit())
			} else {
				require.False(t, test.wantEnabled) // If there is no limit make sure it's expected to be disabled
			}

			code, err := userResolver.CodyGatewayAccess().CodeCompletionsRateLimit(adminContext)
			require.NoError(t, err)
			if chat != nil {
				require.Equal(t, test.wantCode, code.Limit())
			} else {
				require.False(t, test.wantEnabled) // If there is no limit make sure it's expected to be disabled
			}

			assert.Equal(t, test.wantEnabled, userResolver.CodyGatewayAccess().Enabled())
		})
	}
}

func TestCodyGatewayDotcomUserResolverRequestAccess(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(logtest.Scoped(t), dbtest.NewDB(logtest.Scoped(t), t))

	// Admin
	adminUser, err := db.Users().Create(ctx, database.NewUser{Username: "admin", EmailIsVerified: true, Email: "admin@test.com"})
	require.NoError(t, err)

	// Not Admin
	notAdminUser, err := db.Users().Create(ctx, database.NewUser{Username: "verified", EmailIsVerified: true, Email: "verified@test.com"})
	require.NoError(t, err)

	// cody user
	coydUser, err := db.Users().Create(ctx, database.NewUser{Username: "cody", EmailIsVerified: true, Email: "cody@test.com"})
	require.NoError(t, err)
	// Generate a token for the cody user
	_, codyUserToken, err := db.AccessTokens().Create(context.Background(), coydUser.ID, []string{authz.ScopeUserAll}, "cody", coydUser.ID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		user    *types.User
		wantErr error
	}{
		{
			name:    "admin user",
			user:    adminUser,
			wantErr: nil,
		},
		{
			name:    "not admin user",
			user:    notAdminUser,
			wantErr: auth.ErrMustBeSiteAdmin,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Create a request context from the user
			requestContext := actor.WithActor(context.Background(), actor.FromActualUser(test.user))

			// Make request from the test user
			r := productsubscription.CodyGatewayDotcomUserResolver{DB: db}
			_, err := r.DotcomCodyGatewayUserByToken(requestContext, &graphqlbackend.CodyGatewayUsersByAccessTokenArgs{Token: codyUserToken})

			require.ErrorIs(t, err, test.wantErr)

		})
	}
}

func iPtr(i int) *int {
	return &i
}
