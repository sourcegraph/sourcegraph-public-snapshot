// +build pgsqltest

package pgsql

import (
	"reflect"
	"sort"
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

func userLogins(users []*sourcegraph.User) []string {
	var logins []string
	for _, user := range users {
		logins = append(logins, user.Login)
	}
	sort.Strings(logins)
	return logins
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

// TestUsers_Get_existingByBothOnlyOneExist tests the behavior of
// Users.Get when called with both a login and UID, but only one of
// those points to an existing user.
func TestUsers_Get_existingByBothOnlyOneExist(t *testing.T) {
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

	if _, err := s.Get(ctx, sourcegraph.UserSpec{UID: 123, Login: "u"}); !isUserNotFound(err) {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, sourcegraph.UserSpec{UID: created.UID, Login: "doesntexist"}); !isUserNotFound(err) {
		t.Fatal(err)
	}
}

// TestUsers_Get_nonexistentLogin tests the behavior of Users.Get when
// called with a login of a user that does not exist.
func TestUsers_Get_nonexistentLogin(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	user, err := s.Get(ctx, sourcegraph.UserSpec{Login: "doesntexist"})
	if !isUserNotFound(err) {
		t.Fatal(err)
	}
	if user != nil {
		t.Error("user != nil")
	}
}

// TestUsers_Get_nonexistentUID tests the behavior of Users.Get when
// called with a UID of a user that does not exist.
func TestUsers_Get_nonexistentUID(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	user, err := s.Get(ctx, sourcegraph.UserSpec{UID: 456 /* doesn't exist */})
	if !isUserNotFound(err) {
		t.Fatal(err)
	}
	if user != nil {
		t.Error("user != nil")
	}
}

// TestUsers_List_ok tests the behavior of Users.List.
func TestUsers_List_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &users{}
	createUser := newCreateUserFunc(ctx)
	if _, err := createUser(sourcegraph.User{Login: "u0"}); err != nil {
		t.Fatal(err)
	}
	if _, err := createUser(sourcegraph.User{Login: "u1"}); err != nil {
		t.Fatal(err)
	}

	users, err := s.List(ctx, &sourcegraph.UsersListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := 2; len(users) != want {
		t.Errorf("got len(users) == %d, want %d", len(users), want)
	}
	logins := userLogins(users)
	if want := []string{"u0", "u1"}; !reflect.DeepEqual(logins, want) {
		t.Errorf("got logins == %v, want %v", logins, want)
	}
}

func TestUsers_List_query(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Users_List_query(ctx, t, &users{}, newCreateUserFunc(ctx))
}
