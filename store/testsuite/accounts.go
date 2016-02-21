package testsuite

import (
	"regexp"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

type GetAccountFunc func(sourcegraph.UserSpec) (*sourcegraph.User, error)

// Accounts_Create_noLogin tests the behavior of Accounts.Create when
// called with an empty login.
func Accounts_Create_noLogin(ctx context.Context, t *testing.T, s store.Accounts) {
	if _, err := s.Create(ctx, &sourcegraph.User{Login: ""}); err == nil {
		t.Fatal("err == nil")
	}
}

// Accounts_Create_uidAlreadySet tests the behavior of Accounts.Create
// when called with an already populated UID.
func Accounts_Create_uidAlreadySet(ctx context.Context, t *testing.T, s store.Accounts) {
	if _, err := s.Create(ctx, &sourcegraph.User{UID: 123, Login: "u"}); err == nil {
		t.Fatal("err == nil")
	}
}

// Accounts_RequestPasswordReset tests that we can request a password reset. It
// is also used to set up the ResetPassword tests.
func Accounts_RequestPasswordReset(ctx context.Context, t *testing.T, s store.Accounts) {
	u := &sourcegraph.User{UID: 123}
	token, err := s.RequestPasswordReset(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	p := "[0-9a-zA-Z]{44}"
	r := regexp.MustCompile(p)
	if !r.MatchString(token.Token) {
		t.Errorf("token should match %s", p)
	}
}

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
