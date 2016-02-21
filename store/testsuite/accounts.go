package testsuite

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

type GetAccountFunc func(sourcegraph.UserSpec) (*sourcegraph.User, error)

// Accounts_ResetPassword_ok tests that we can successfully reset a password.
func Accounts_ResetPassword_ok(ctx context.Context, t *testing.T, s store.Accounts) {
	u := &sourcegraph.User{UID: 123}
	token, err := s.RequestPasswordReset(ctx, u)
	if err != nil {
		t.Fatal(err)
	}

	newPass := &sourcegraph.NewPassword{Password: "a", Token: &sourcegraph.PasswordResetToken{Token: token.Token}}
	if err := s.ResetPassword(ctx, newPass); err != nil {
		t.Fatal(err)
	}
}

// Accounts_ResetPassword_badtoken tests that we cannot reset a password without
// the correct token.
func Accounts_ResetPassword_badtoken(ctx context.Context, t *testing.T, s store.Accounts) {
	newPass := &sourcegraph.NewPassword{Password: "a", Token: &sourcegraph.PasswordResetToken{Token: "b"}}
	if err := s.ResetPassword(ctx, newPass); err == nil {
		t.Errorf("Should have gotten error reseting password, got nil instead")
	}
}
