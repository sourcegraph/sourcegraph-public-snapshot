package fs

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store/testsuite"
)

func newGetAccountFunc(ctx context.Context) testsuite.GetAccountFunc {
	return func(user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		return (&Users{}).Get(ctx, user)
	}
}

func TestAccounts_Create_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_ok(ctx, t, &Accounts{}, newGetAccountFunc(ctx))
}

func TestAccounts_Create_duplicate(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_duplicate(ctx, t, &Accounts{})
}

func TestAccounts_Create_noLogin(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_noLogin(ctx, t, &Accounts{})
}

func TestAccounts_Create_uidAlreadySet(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Accounts_Create_uidAlreadySet(ctx, t, &Accounts{})
}
