// +build pgsqltest

package localstore

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

// TestAccounts_Create_ok tests the behavior of Accounts.Create when
// called with correct args.
func TestAccounts_Create_ok(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	getAccount := func(user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		return (&users{}).Get(ctx, user)
	}

	want := sourcegraph.User{Login: "u", Name: "n"}
	email := &sourcegraph.EmailAddr{Email: "email@email.email"}

	created, err := s.Create(ctx, &want, email)
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

// TestAccounts_Create_duplicate tests the behavior of Accounts.Create
// when called with an existing (duplicate) login.
func TestAccounts_Create_duplicate(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	email := &sourcegraph.EmailAddr{Email: "email@email.email"}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: "u"}, email); err != nil {
		t.Fatal(err)
	}

	email = &sourcegraph.EmailAddr{Email: "email1@email.email"}

	_, err := s.Create(ctx, &sourcegraph.User{Login: "u"}, email)
	if _, ok := err.(*store.AccountAlreadyExistsError); !ok {
		t.Fatalf("got err type %T, want %T", err, &store.AccountAlreadyExistsError{})
	}
}

// TestAccounts_Create_noLogin tests the behavior of Accounts.Create when
// called with an empty login.
func TestAccounts_Create_noLogin(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	email := &sourcegraph.EmailAddr{Email: "email@email.email"}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: ""}, email); err == nil {
		t.Fatal("err == nil")
	}
}

// TestAccounts_Create_uidAlreadySet tests the behavior of Accounts.Create
// when called with an already populated UID.
func TestAccounts_Create_uidAlreadySet(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	email := &sourcegraph.EmailAddr{Email: "email@email.email"}

	if _, err := s.Create(ctx, &sourcegraph.User{UID: 123, Login: "u"}, email); err == nil {
		t.Fatal("err == nil")
	}
}

// TestAccounts_Create_noEmail tests the behavior of Accounts.Create
// when called with an empty email
func TestAccounts_Create_noEmail(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	email := &sourcegraph.EmailAddr{Email: ""}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: ""}, email); err == nil {
		t.Fatal("err == nil")
	}
}

// TestAccounts_Create_ExistingEmail tests the behavior of Accounts.Create
// when called with an email that is used by another account
func TestAccounts_Create_ExistingEmail(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	email := &sourcegraph.EmailAddr{Email: "email@email.email"}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: "u"}, email); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: "u2"}, email); err != nil {
		if !strings.Contains(err.Error(), "has already been registered with another account") {
			t.Fatal("wrong error was produced, was expecing code == 6 and error about email already taken")
		}
	}
}

// TestAccounts_RequestPasswordReset tests that we can request a password
// reset. It is also used to set up the ResetPassword tests.
func TestAccounts_RequestPasswordReset(t *testing.T) {
	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
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

// TestAccounts_ResetPassword_ok tests that we can successfully reset a password.
func TestAccounts_ResetPassword_ok(t *testing.T) {
	// t.Parallel() // TODO s.RequestPasswordReset occasionally has a data race with the same function from TestAccounts_RequestPasswordReset
	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
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

// TestAccounts_ResetPassword_badtoken tests that we cannot reset a password without
// the correct token.
func TestAccounts_ResetPassword_badtoken(t *testing.T) {
	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := accounts{}
	newPass := &sourcegraph.NewPassword{Password: "a", Token: &sourcegraph.PasswordResetToken{Token: "b"}}
	if err := s.ResetPassword(ctx, newPass); err == nil {
		t.Errorf("Should have gotten error reseting password, got nil instead")
	}
}
