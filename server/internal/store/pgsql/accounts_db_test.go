// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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

func TestAccounts_Create_duplicate(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_duplicate(ctx, t, &accounts{})
}

func TestAccounts_Create_noLogin(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_noLogin(ctx, t, &accounts{})
}

func TestAccounts_Create_uidAlreadySet(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_uidAlreadySet(ctx, t, &accounts{})
}

func TestAccounts_RequestPasswordReset(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_RequestPasswordReset(ctx, t, &accounts{})
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
