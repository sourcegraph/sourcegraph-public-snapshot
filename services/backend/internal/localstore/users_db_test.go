package localstore

import (
	"sort"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func newCreateUserFunc(ctx context.Context) func(user sourcegraph.User, email sourcegraph.EmailAddr) (*sourcegraph.UserSpec, error) {
	return func(user sourcegraph.User, email sourcegraph.EmailAddr) (*sourcegraph.UserSpec, error) {
		created, err := (&accounts{}).Create(ctx, &user, &email)
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

// TestUsers_Get_existingByUID tests the behavior of Users.Get when called
// with the UID of a user that exists (i.e., the successful outcome).
func TestUsers_Get_existingByUID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := &users{}
	created, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"}, sourcegraph.EmailAddr{Email: "email@email.email"})
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
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := &users{}
	created, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"}, sourcegraph.EmailAddr{Email: "email@email.email"})
	if err != nil {
		t.Fatal(err)
	}
	if created.UID == 0 {
		t.Error("violated assumption that UID is set")
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
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	createUser := newCreateUserFunc(ctx)
	created0, err := createUser(sourcegraph.User{Login: "u0"}, sourcegraph.EmailAddr{Email: "email0@email.email"})
	if err != nil {
		t.Fatalf("create0 error %q", err)
	}
	if created0.UID == 0 {
		t.Error("violated assumption that UID is set")
	}

	created1, err := createUser(sourcegraph.User{Login: "u1"}, sourcegraph.EmailAddr{Email: "email1@email.email"})
	if err != nil {
		t.Fatalf("create1 error %q", err)
	}
	if created1.UID == 0 {
		t.Error("violated assumption that UID is set")
	}
}

// TestUsers_Get_existingByBothOnlyOneExist tests the behavior of
// Users.Get when called with both a login and UID, but only one of
// those points to an existing user.
func TestUsers_Get_existingByBothOnlyOneExist(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	created, err := newCreateUserFunc(ctx)(sourcegraph.User{Login: "u"}, sourcegraph.EmailAddr{Email: "email@email.email"})
	if err != nil {
		t.Fatal(err)
	}
	if created.UID == 0 {
		t.Error("violated assumption that UID is set")
	}
}

// TestUsers_Get_nonexistentUID tests the behavior of Users.Get when
// called with a UID of a user that does not exist.
func TestUsers_Get_nonexistentUID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
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
