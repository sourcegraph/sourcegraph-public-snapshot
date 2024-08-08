package openidconnect

import (
	"context"
	"strings"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAllowSignup(t *testing.T) {
	allow := true
	disallow := false
	tests := map[string]struct {
		allowSignup          *bool
		usernamePrefix       string
		shouldAllowSignup    bool
		additionalProperties telemetry.EventMetadata
	}{
		"nil": {
			allowSignup:       nil,
			shouldAllowSignup: true,
		},
		"true": {
			allowSignup:       &allow,
			shouldAllowSignup: true,
		},
		"false": {
			allowSignup:       &disallow,
			shouldAllowSignup: false,
		},
		"with username prefix": {
			allowSignup:       &disallow,
			shouldAllowSignup: false,
			usernamePrefix:    "sourcegraph-operator-",
		},
		"with metadata": {
			allowSignup:          &allow,
			shouldAllowSignup:    true,
			additionalProperties: telemetry.EventMetadata{"foo": 1},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
				require.Equal(t, test.shouldAllowSignup, op.CreateIfNotExist)
				require.True(
					t,
					strings.HasPrefix(op.UserProps.Username, test.usernamePrefix),
					"The username %q does not have prefix %q", op.UserProps.Username, test.usernamePrefix,
				)
				return false, 0, "", nil
			}
			db := dbmocks.NewStrictMockDB()
			_ = telemetrytest.AddDBMocks(db)

			_, _, _, err := getOrCreateUser(
				context.Background(),
				logtest.Scoped(t),
				db,
				schema.OpenIDConnectAuthProvider{
					ClientID:           testClientID,
					ClientSecret:       "aaaaaaaaaaaaaaaaaaaaaaaaa",
					RequireEmailDomain: "example.com",
					AllowSignup:        test.allowSignup,
				},
				&oauth2.Token{},
				&oidc.IDToken{},
				&oidc.UserInfo{
					Email:         "foo@bar.com",
					EmailVerified: true,
				},
				&userClaims{},
				test.usernamePrefix,
				test.additionalProperties,

				&hubspot.ContactProperties{
					AnonymousUserID:            "anonymous-user-id-123",
					FirstSourceURL:             "https://example.com/",
					LastSourceURL:              "https://example.com/",
					LastPageSeenShort:          "https://example.com/",
					LastPageSeenMid:            "https://example.com/",
					LastPageSeenLong:           "https://example.com/",
					MostRecentReferrerUrl:      "https://example.com/",
					MostRecentReferrerUrlShort: "https://example.com/",
					MostRecentReferrerUrlMid:   "https://example.com/",
					MostRecentReferrerUrlLong:  "https://example.com/",
					SessionUTMCampaign:         "session-utm-campaign-123",
					UtmCampaignShort:           "utm-campaign-short-123",
					UtmCampaignMid:             "utm-campaign-mid-123",
					UtmCampaignLong:            "utm-campaign-long-123",
					SessionUTMSource:           "session-utm-source-123",
					UtmSourceShort:             "utm-source-short-123",
					UtmSourceMid:               "utm-source-mid-123",
					UtmSourceLong:              "utm-source-long-123",
					SessionUTMMedium:           "session-utm-medium-123",
					UtmMediumShort:             "utm-medium-short-123",
					UtmMediumMid:               "utm-medium-mid-123",
					UtmMediumLong:              "utm-medium-long-123",
					SessionUTMTerm:             "session-utm-term-123",
					UtmTermShort:               "utm-term-short-123",
					UtmTermMid:                 "utm-term-mid-123",
					UtmTermLong:                "utm-term-long-123",
					SessionUTMContent:          "session-utm-content-123",
					UtmContentShort:            "utm-content-short-123",
					UtmContentMid:              "utm-content-mid-123",
					UtmContentLong:             "utm-content-long-123",
				})
			require.NoError(t, err)
		})
	}
}

func TestGetPublicExternalAccountData(t *testing.T) {
	t.Run("confirm that empty account data does not panic", func(t *testing.T) {
		data := ExternalAccountData{}
		encryptedData, err := encryption.NewUnencryptedJSON[any](data)
		require.NoError(t, err)

		accountData := &extsvc.AccountData{
			Data: encryptedData,
		}

		want := extsvc.PublicAccountData{}

		got, err := GetPublicExternalAccountData(context.Background(), accountData)
		require.NoError(t, err)
		require.Equal(t, want, *got)
	})
}
