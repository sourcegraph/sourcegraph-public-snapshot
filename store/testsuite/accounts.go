package testsuite

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

type GetAccountFunc func(sourcegraph.UserSpec) (*sourcegraph.User, error)

// Accounts_Create_ok tests the behavior of Accounts.Create when
// called with correct args.
func Accounts_Create_ok(ctx context.Context, t *testing.T, s store.Accounts, getAccount GetAccountFunc) {
	want := sourcegraph.User{Login: "u", Name: "n"}

	created, err := s.Create(ctx, &want)
	if err != nil {
		t.Fatal(err)
	}

	if created.Login != want.Login {
		t.Errorf("got Login == %q, want %q", created.Login, want.Login)
	}
	if created.Name != want.Name {
		t.Errorf("got Name == %q, want %q", created.Name, want.Name)
	}

	got, err := getAccount(sourcegraph.UserSpec{Login: "u"})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, created) {
		t.Errorf("Create: got %+v, want %+v", got, created)
	}
}

// Accounts_Create_duplicate tests the behavior of Accounts.Create
// when called with an existing (duplicate) client ID.
func Accounts_Create_duplicate(ctx context.Context, t *testing.T, s store.Accounts) {
	if _, err := s.Create(ctx, &sourcegraph.User{Login: "u"}); err != nil {
		t.Fatal(err)
	}

	_, err := s.Create(ctx, &sourcegraph.User{Login: "u"})
	if _, ok := err.(*store.AccountAlreadyExistsError); !ok {
		t.Fatalf("got err type %T, want %T", err, &store.AccountAlreadyExistsError{})
	}
}

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
		fmt.Errorf("token should match %s", p)
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
