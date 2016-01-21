package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func TestUserSettingsAccess(t *testing.T) {
	origFlag := authutil.ActiveFlags.Source
	authutil.ActiveFlags.Source = "local"
	defer func() {
		authutil.ActiveFlags.Source = origFlag
	}()

	c, mock := apptest.New()

	var calledGet bool
	mock.Users.Get_ = func(context.Context, *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		calledGet = true
		return &sourcegraph.User{UID: 1, Login: "u"}, nil
	}

	var calledOrgsList bool
	mock.Orgs.List_ = func(ctx context.Context, _ *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
		calledOrgsList = true
		return &sourcegraph.OrgList{}, nil
	}

	// Test that the settings page is accessible as the same user.
	mock.Ctx = handlerutil.WithUser(mock.Ctx, &sourcegraph.UserSpec{
		UID:   1,
		Login: "u",
	})

	if _, err := c.GetOK(router.Rel.URLToUserSubroute(router.UserSettingsProfile, "u").String()); err != nil {
		t.Fatalf("expected to succeed when accessing user's own settings page. got %v", err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}

	if !calledOrgsList {
		t.Error("!calledOrgsList")
	}

	// Test that the settings page is not accessible as a different user.
	mock.Ctx = handlerutil.WithUser(mock.Ctx, &sourcegraph.UserSpec{
		UID:   2,
		Login: "w",
	})
	calledGet = false

	if _, err := c.GetOK(router.Rel.URLToUserSubroute(router.UserSettingsProfile, "u").String()); err == nil {
		t.Fatalf("expected to get error when accessing another user's settings page. got nil")
	}

	if !calledGet {
		t.Error("!calledGet")
	}

	// Test that the settings page is not accessible as an anonymous user.
	mock.Ctx = handlerutil.WithUser(mock.Ctx, nil)

	if _, err := c.GetOK(router.Rel.URLToUserSubroute(router.UserSettingsProfile, "u").String()); err == nil {
		t.Fatalf("expected to get error when accessing a user's settings page as anonymous. got nil")
	}
}
