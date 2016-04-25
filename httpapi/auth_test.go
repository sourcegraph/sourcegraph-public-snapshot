package httpapi

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestAuthInfo_includeUser(t *testing.T) {
	c, mock := newTest()

	want := &authInfo{AuthInfo: sourcegraph.AuthInfo{UID: 1}, IncludedUser: &sourcegraph.User{UID: 1}}

	var calledIdentify bool
	mock.Auth.Identify_ = func(context.Context, *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
		calledIdentify = true
		return &want.AuthInfo, nil
	}
	calledUsersGet := mock.Users.MockGetByUID(t, 1)

	var authInfo *authInfo
	if err := c.GetJSON("/auth-info", &authInfo); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(authInfo, want) {
		t.Errorf("got %+v, want %+v", authInfo, want)
	}
	if !calledIdentify {
		t.Error("!calledIdentify")
	}
	if !*calledUsersGet {
		t.Error("!calledUsersGet")
	}
}

func TestAuthInfo_noUser(t *testing.T) {
	c, mock := newTest()

	want := &authInfo{AuthInfo: sourcegraph.AuthInfo{UID: 0}, IncludedUser: nil}

	var calledIdentify bool
	mock.Auth.Identify_ = func(context.Context, *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
		calledIdentify = true
		return &want.AuthInfo, nil
	}

	var authInfo *authInfo
	if err := c.GetJSON("/auth-info", &authInfo); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(authInfo, want) {
		t.Errorf("got %+v, want %+v", authInfo, want)
	}
	if !calledIdentify {
		t.Error("!calledIdentify")
	}
}
