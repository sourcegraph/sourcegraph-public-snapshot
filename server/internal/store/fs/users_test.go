package fs

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store/testsuite"
)

func newCreateUserFunc(ctx context.Context) testsuite.CreateUserFunc {
	return func(user sourcegraph.User) (*sourcegraph.UserSpec, error) {
		created, err := (&Accounts{}).Create(ctx, &user)
		if err != nil {
			return nil, err
		}
		spec := created.Spec()
		return &spec, nil
	}
}

func TestUsers_Get_existingByLogin(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByLogin(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByUID(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByUID(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBoth(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBoth(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBothConflict(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBothConflict(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBothOnlyOneExist(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBothOnlyOneExist(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_nonexistentLogin(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_nonexistentLogin(ctx, t, &Users{})
}

func TestUsers_Get_nonexistentUID(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_nonexistentUID(ctx, t, &Users{})
}

func TestUsers_List_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_List_ok(ctx, t, &Users{}, newCreateUserFunc(ctx))
}

func TestUsers_List_query(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Users_List_query(ctx, t, &Users{}, newCreateUserFunc(ctx))
}
