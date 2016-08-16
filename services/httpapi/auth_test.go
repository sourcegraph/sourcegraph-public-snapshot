package httpapi

import (
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestAuthInfo_includeUserAndEmails(t *testing.T) {
	c, mock := newTest()

	want := &authInfo{
		AuthInfo:       sourcegraph.AuthInfo{UID: 1},
		IncludedUser:   &sourcegraph.User{UID: 1},
		IncludedEmails: []*sourcegraph.EmailAddr{{Email: "a@a.com", Primary: true}},
		GitHubToken:    &sourcegraph.ExternalToken{Token: "t"},
	}

	var calledIdentify, calledAuthGetExternalToken bool
	mock.Auth.Identify_ = func(context.Context, *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
		calledIdentify = true
		return &want.AuthInfo, nil
	}
	mock.Auth.GetExternalToken_ = func(context.Context, *sourcegraph.ExternalTokenSpec) (*sourcegraph.ExternalToken, error) {
		calledAuthGetExternalToken = true
		return want.GitHubToken, nil
	}
	calledUsersGet := mock.Users.MockGetByUID(t, 1)
	calledUsersListEmails := mock.Users.MockListEmails(t, "a@a.com")

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
	if !calledAuthGetExternalToken {
		t.Error("!calledAuthGetExternalToken")
	}
	if !*calledUsersGet {
		t.Error("!calledUsersGet")
	}
	if !*calledUsersListEmails {
		t.Error("!calledUsersListEmails")
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
