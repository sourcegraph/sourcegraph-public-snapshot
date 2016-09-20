package localstore

import (
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// TestAccounts_Create_ok tests the behavior of Accounts.Create when
// called with correct args.
func TestAccounts_Create_ok(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &accounts{}

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
}

// TestAccounts_Create_duplicate tests the behavior of Accounts.Create
// when called with an existing (duplicate) login.
func TestAccounts_Create_duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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
	if grpc.Code(err) != codes.AlreadyExists {
		t.Fatalf("got err code %d, want %d", grpc.Code(err), codes.AlreadyExists)
	}
}

// TestAccounts_Create_noLogin tests the behavior of Accounts.Create when
// called with an empty login.
func TestAccounts_Create_noLogin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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
	if testing.Short() {
		t.Skip()
	}

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
	if testing.Short() {
		t.Skip()
	}

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
	if testing.Short() {
		t.Skip()
	}

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
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	before := time.Now()
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

	req, err := unmarshalResetRequest(ctx, token.Token)
	if err != nil {
		t.Fatal(err)
	}
	if !req.ExpiresAt.After(before) {
		t.Errorf("token's expiration date: %v should be at some point in the future", req.ExpiresAt)
	}
}

// TestAccounts_ResetPassword_ok tests that we can successfully reset a password.
func TestAccounts_ResetPassword_ok(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// t.Parallel() // TODO s.RequestPasswordReset occasionally has a data race with the same function from TestAccounts_RequestPasswordReset
	ctx, _, done := testContext()
	defer done()

	s := &accounts{}
	u := &sourcegraph.User{UID: 123}
	token, err := s.RequestPasswordReset(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	oldReq, err := unmarshalResetRequest(ctx, token.Token)
	if err != nil {
		t.Fatal(err)
	}
	newPass := &sourcegraph.NewPassword{Password: "a", Token: &sourcegraph.PasswordResetToken{Token: token.Token}}
	if err := s.ResetPassword(ctx, newPass); err != nil {
		t.Fatal(err)
	}

	newReq, err := unmarshalResetRequest(ctx, newPass.Token.Token)
	if err != nil {
		t.Fatal(err)
	}
	if newReq.ExpiresAt.After(oldReq.ExpiresAt) {
		t.Errorf("token's new expiration date: %v should be before its previous expiration date: %v", newReq.ExpiresAt, oldReq.ExpiresAt)
	}
}

// TestAccounts_ResetPassword_badtoken tests that we cannot reset a password without
// the correct token.
func TestAccounts_ResetPassword_badtoken(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := accounts{}
	newPass := &sourcegraph.NewPassword{Password: "a", Token: &sourcegraph.PasswordResetToken{Token: "b"}}
	if err := s.ResetPassword(ctx, newPass); err == nil {
		t.Errorf("should have gotten error reseting password, got nil instead")
	}
}

// TestAccounts_CleanExpiredResets tests that expired password reset requests are removed from the
// database whenever cleanExpiredResets is called.
func TestAccounts_CleanExpiredResets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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
	s.cleanExpiredResets(ctx)
	_, err = unmarshalResetRequest(ctx, newPass.Token.Token)
	if err != sql.ErrNoRows {
		t.Fatalf("got this error: %s, should have gotten just a NoRow error instead when trying to retrieve a cleaned reset request", err)
	}
}
