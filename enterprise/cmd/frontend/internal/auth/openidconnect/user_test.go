package openidconnect

import (
	"context"
	"testing"

	"github.com/coreos/go-oidc"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAllowSignup(t *testing.T) {
	allow := true
	disallow := false
	tests := map[string]struct {
		allowSignup       *bool
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
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
				if test.shouldAllowSignup != op.CreateIfNotExist {
					t.Fatalf("op.CreateIfNotExist: want %v got %v\n", test.shouldAllowSignup, op.CreateIfNotExist)
				}
				return 0, "", nil
			}
			p := &provider{config: schema.OpenIDConnectAuthProvider{
				ClientID:           testClientID,
				ClientSecret:       "aaaaaaaaaaaaaaaaaaaaaaaaa",
				RequireEmailDomain: "example.com",
				AllowSignup:        test.allowSignup,
			}, oidc: &oidcProvider{}}
			_, _, err := getOrCreateUser(context.Background(), database.NewStrictMockDB(), p, &oidc.IDToken{}, &oidc.UserInfo{Email: "foo@bar.com", EmailVerified: true}, &userClaims{})
			if err != nil {
				t.Errorf("err: expected nil, got %v\n", err)
			}
		})
	}
}
