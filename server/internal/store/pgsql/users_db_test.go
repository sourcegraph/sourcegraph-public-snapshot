// +build pgsqltest

package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
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

func isUserNotFound(err error) bool {
	_, ok := err.(*store.UserNotFoundError)
	return ok
}

// TestUsers_Get_existingByLogin tests the behavior of Users.Get when
// called with the login of a user that exists (i.e., the successful
// outcome).
func TestUsers_Get_existingByLogin(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	if _, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"}); err != nil {
		t.Fatal(err)
	}

	user, err := s.Get(ctx, sourcegraph.UserSpec{Login: "u"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Login != want {
		t.Errorf("got login == %q, want %q", user.Login, want)
	}
}

// TestUsers_Get_existingByUID tests the behavior of Users.Get when called
// with the UID of a user that exists (i.e., the successful outcome).
func TestUsers_Get_existingByUID(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	created, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"})
	if err != nil {
		t.Fatal(err)
	}

	user, err := s.Get(ctx, sourcegraph.UserSpec{UID: created.UID})
	if err != nil {
		t.Fatal(err)
	}
	if user.Spec() != *created {
		t.Errorf("got user spec == %+v, want %+v", user.Spec(), *created)
	}
}

// TestUsers_Get_existingByBoth tests the behavior of Users.Get when
// called with both the login and UID of a user that exists (i.e., the
// successful outcome).
func TestUsers_Get_existingByBoth(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	created, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Login == "" || created.UID == 0 {
		t.Error("violated assumption that both login and UID are set")
	}

	user, err := s.Get(ctx, *created)
	if err != nil {
		t.Fatal(err)
	}
	if user.Spec() != *created {
		t.Errorf("got user spec == %+v, want %+v", user.Spec(), *created)
	}
}

// TestUsers_Get_existingByBothConflict tests the behavior of Users.Get
// when called with both a login and UID, but when those do not both
// refer to the same user.
func TestUsers_Get_existingByBothConflict(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	createUser := newCreateUserFunc(ctx)
	created0, err := createUser(sourcegraph.User{Login: "u0"})
	if err != nil {
		t.Fatal(err)
	}
	if created0.Login == "" || created0.UID == 0 {
		t.Error("violated assumption that both login and UID are set")
	}

	created1, err := createUser(sourcegraph.User{Login: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if created1.Login == "" || created1.UID == 0 {
		t.Error("violated assumption that both login and UID are set")
	}

	if _, err := s.Get(ctx, sourcegraph.UserSpec{UID: created0.UID, Login: created1.Login}); !isUserNotFound(err) {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, sourcegraph.UserSpec{UID: created1.UID, Login: created0.Login}); !isUserNotFound(err) {
		t.Fatal(err)
	}
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
