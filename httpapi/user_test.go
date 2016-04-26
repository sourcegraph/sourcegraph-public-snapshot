package httpapi

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestUser(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.User{UID: 1}

	calledUsersGet := mock.Users.MockGetByUID(t, 1)

	var user *sourcegraph.User
	if err := c.GetJSON("/users/1$", &user); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user, want) {
		t.Errorf("got %+v, want %+v", user, want)
	}
	if !*calledUsersGet {
		t.Error("!calledUsersGet")
	}
}
