package backend

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestUsersService_Get(t *testing.T) {
	var s users
	ctx, mock := testContext()

	wantUser := &sourcegraph.User{UID: 1, Login: "u"}

	calledGet := mock.stores.Users.MockGet_Return(t, wantUser)

	user, err := s.Get(ctx, &sourcegraph.UserSpec{UID: 1})
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
