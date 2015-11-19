package local

import (
	"reflect"
	"testing"

	authpkg "src.sourcegraph.com/sourcegraph/auth"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestUsersService_Get(t *testing.T) {
	var s users
	ctx, mock := testContext()

	wantUser := &sourcegraph.User{UID: 1, Login: "u"}

	calledGet := mock.stores.Users.MockGet_Return(t, wantUser)

	user, err := s.Get(ctx, &sourcegraph.UserSpec{UID: 1, Login: "u"})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !reflect.DeepEqual(user, wantUser) {
		t.Errorf("got %+v, want %+v", user, wantUser)
	}
}

func TestUsersService_List(t *testing.T) {
	var s users
	ctx, mock := testContext()

	wantUsers := &sourcegraph.UserList{
		Users: []*sourcegraph.User{
			{UID: 1, Login: "u1"},
			{UID: 2, Login: "u2"},
		},
	}

	calledList := mock.stores.Users.MockList(t, "u1", "u2")

	users, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !reflect.DeepEqual(users, wantUsers) {
		t.Errorf("got %+v, want %+v", users, wantUsers)
	}
}

func TestVerifyCanReadOwnEmail(t *testing.T) {
	var s users
	ctx, _ := testContext()

	actor := sourcegraph.UserSpec{UID: 100}
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: int(actor.UID)})
	if err := s.verifyCanReadEmail(ctx, actor); err != nil {
		t.Error("Should be able to read own email")
	}

	if err := s.verifyCanReadEmail(ctx, sourcegraph.UserSpec{UID: 123}); err == nil {
		t.Error("Should not be allowed to read other user's email")
	}
}
