package testsuite

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sort"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// CreateUserFunc is used by Users_Get_* tests to create test users.
type CreateUserFunc func(sourcegraph.User) (*sourcegraph.UserSpec, error)

// Users_Get_existingByLogin tests the behavior of Users.Get when
// called with the login of a user that exists (i.e., the successful
// outcome).
func Users_Get_existingByLogin(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
	if _, err := createUser(sourcegraph.User{Login: "u"}); err != nil {
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

// Users_Get_existingByUID tests the behavior of Users.Get when called
// with the UID of a user that exists (i.e., the successful outcome).
func Users_Get_existingByUID(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
	created, err := createUser(sourcegraph.User{Login: "u"})
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

// Users_Get_existingByBoth tests the behavior of Users.Get when
// called with both the login and UID of a user that exists (i.e., the
// successful outcome).
func Users_Get_existingByBoth(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
	created, err := createUser(sourcegraph.User{Login: "u"})
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

// Users_Get_existingByBothConflict tests the behavior of Users.Get
// when called with both a login and UID, but when those do not both
// refer to the same user.
func Users_Get_existingByBothConflict(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
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

// Users_Get_existingByBothOnlyOneExist tests the behavior of
// Users.Get when called with both a login and UID, but only one of
// those points to an existing user.
func Users_Get_existingByBothOnlyOneExist(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
	created, err := createUser(sourcegraph.User{Login: "u"})
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

// Users_Get_nonexistentLogin tests the behavior of Users.Get when
// called with a login of a user that does not exist.
func Users_Get_nonexistentLogin(ctx context.Context, t *testing.T, s store.Users) {
	user, err := s.Get(ctx, sourcegraph.UserSpec{Login: "doesntexist"})
	if !isUserNotFound(err) {
		t.Fatal(err)
	}
	if user != nil {
		t.Error("user != nil")
	}
}

// Users_Get_nonexistentUID tests the behavior of Users.Get when
// called with a UID of a user that does not exist.
func Users_Get_nonexistentUID(ctx context.Context, t *testing.T, s store.Users) {
	user, err := s.Get(ctx, sourcegraph.UserSpec{UID: 456 /* doesn't exist */})
	if !isUserNotFound(err) {
		t.Fatal(err)
	}
	if user != nil {
		t.Error("user != nil")
	}
}

// Users_List_ok tests the behavior of Users.List.
func Users_List_ok(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
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

// Users_List_query tests the behavior of Users.List when called with
// a query.
func Users_List_query(ctx context.Context, t *testing.T, s store.Users, createUser CreateUserFunc) {
	if _, err := createUser(sourcegraph.User{Login: "u0"}); err != nil {
		t.Fatal(err)
	}
	if _, err := createUser(sourcegraph.User{Login: "u1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := createUser(sourcegraph.User{Login: "u12"}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		query string
		want  []string
	}{
		{"u", []string{"u0", "u1", "u12"}},
		{"u1", []string{"u1", "u12"}},
		{"u12", []string{"u12"}},
		{"u9", nil},
		{"U1", []string{"u1", "u12"}},
	}
	for _, test := range tests {
		users, err := s.List(ctx, &sourcegraph.UsersListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := userLogins(users); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got users %v, want %v", test.query, got, test.want)
		}
	}
}

func userLogins(users []*sourcegraph.User) []string {
	var logins []string
	for _, user := range users {
		logins = append(logins, user.Login)
	}
	sort.Strings(logins)
	return logins
}

func isUserNotFound(err error) bool {
	_, ok := err.(*store.UserNotFoundError)
	return ok
}
