// +build pgsqltest

package pgsql

import (
	"reflect"
	"regexp"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func newGetAccountFunc(ctx context.Context) testsuite.GetAccountFunc {
	return func(user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		return (&users{}).Get(ctx, user)
	}
}

// TestAccounts_Create_ok tests the behavior of Accounts.Create when
// called with correct args.
func TestAccounts_Create_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s, getAccount := &accounts{}, newGetAccountFunc(ctx)
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

// TestAccounts_Create_duplicate tests the behavior of Accounts.Create
// when called with an existing (duplicate) client ID.
func TestAccounts_Create_duplicate(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &accounts{}

	if _, err := s.Create(ctx, &sourcegraph.User{Login: "u"}); err != nil {
		t.Fatal(err)
	}

	_, err := s.Create(ctx, &sourcegraph.User{Login: "u"})
	if _, ok := err.(*store.AccountAlreadyExistsError); !ok {
		t.Fatalf("got err type %T, want %T", err, &store.AccountAlreadyExistsError{})
	}
}

// TestAccounts_Create_noLogin tests the behavior of Accounts.Create when
// called with an empty login.
func TestAccounts_Create_noLogin(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &accounts{}
	if _, err := s.Create(ctx, &sourcegraph.User{Login: ""}); err == nil {
		t.Fatal("err == nil")
	}
}

// TestAccounts_Create_uidAlreadySet tests the behavior of Accounts.Create
// when called with an already populated UID.
func TestAccounts_Create_uidAlreadySet(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &accounts{}
	if _, err := s.Create(ctx, &sourcegraph.User{UID: 123, Login: "u"}); err == nil {
		t.Fatal("err == nil")
	}
}

// TestAccounts_RequestPasswordReset tests that we can request a password
// reset. It is also used to set up the ResetPassword tests.
func TestAccounts_RequestPasswordReset(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
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

func TestAccounts_ResetPassword_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_ResetPassword_ok(ctx, t, &accounts{})
}

func TestAccounts_ResetPassword_badtoken(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_ResetPassword_badtoken(ctx, t, &accounts{})
}
