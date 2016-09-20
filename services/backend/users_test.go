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
