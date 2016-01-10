// +build pgsqltest

package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func newCreateUserFunc(ctx context.Context) testsuite.CreateUserFunc {
	return func(user sourcegraph.User) (*sourcegraph.UserSpec, error) {
		created, err := (&accounts{}).Create(ctx, &user)
		if err != nil {
			return nil, err
		}
		spec := created.Spec()
		return &spec, nil
	}
}

func TestUsers_Get_existingByLogin(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByLogin(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByUID(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByUID(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBoth(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBoth(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBothConflict(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBothConflict(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_existingByBothOnlyOneExist(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_existingByBothOnlyOneExist(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_Get_nonexistentLogin(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_nonexistentLogin(ctx, t, &users{})
}

func TestUsers_Get_nonexistentUID(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_Get_nonexistentUID(ctx, t, &users{})
}

func TestUsers_List_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_List_ok(ctx, t, &users{}, newCreateUserFunc(ctx))
}

func TestUsers_List_query(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_List_query(ctx, t, &users{}, newCreateUserFunc(ctx))
}
