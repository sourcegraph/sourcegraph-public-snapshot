package openidconnect

import (
	"context"
	"strings"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAllowSignup(t *testing.T) {
	allow := true
	disallow := false
	tests := map[string]struct {
		allowSignup       *bool
		usernamePrefix    string
		shouldAllowSignup bool
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
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
				require.Equal(t, test.shouldAllowSignup, op.CreateIfNotExist)
				require.True(
					t,
					strings.HasPrefix(op.UserProps.Username, test.usernamePrefix),
					"The username %q does not have prefix %q", op.UserProps.Username, test.usernamePrefix,
				)
				return 0, "", nil
			}
			p := &Provider{
				config: schema.OpenIDConnectAuthProvider{
					ClientID:           testClientID,
					ClientSecret:       "aaaaaaaaaaaaaaaaaaaaaaaaa",
					RequireEmailDomain: "example.com",
					AllowSignup:        test.allowSignup,
				},
				oidc: &oidcProvider{},
			}
			_, _, err := getOrCreateUser(
				context.Background(),
				dbmocks.NewStrictMockDB(),
				p,
				&oidc.IDToken{},
				&oidc.UserInfo{
					Email:         "foo@bar.com",
					EmailVerified: true,
				},
				&userClaims{},
				test.usernamePrefix,
			)
			require.NoError(t, err)
		})
	}
}
