package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func TestUserSettingsAccess(t *testing.T) {
	origSource := authutil.ActiveFlags.Source
	defer func() {
		authutil.ActiveFlags.Source = origSource
	}()
	authutil.ActiveFlags.Source = "local"

	c, mock := apptest.New()

	var calledGet bool
	mock.Users.Get_ = func(context.Context, *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		calledGet = true
		return &sourcegraph.User{UID: 1, Login: "u"}, nil
	}

	var calledListEmails bool
	mock.Users.ListEmails_ = func(ctx context.Context, in *sourcegraph.UserSpec) (*sourcegraph.EmailAddrList, error) {
		calledListEmails = true
		return &sourcegraph.EmailAddrList{}, nil
	}

	var calledOrgsList bool
	mock.Orgs.List_ = func(ctx context.Context, _ *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
		calledOrgsList = true
		return &sourcegraph.OrgList{}, nil
	}

	// Test that the settings page is accessible as the same user.
	mock.Ctx = handlerutil.WithUser(mock.Ctx, sourcegraph.UserSpec{
		UID:   1,
		Login: "u",
	})

	if _, err := c.GetOK(router.Rel.URLToUserSubroute(router.UserSettingsProfile, "u").String()); err != nil {
		t.Fatalf("expected to succeed when accessing user's own settings page. got %v", err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}

	if !calledListEmails {
		t.Error("!calledListEmails")
	}

	if !calledOrgsList {
		t.Error("!calledOrgsList")
	}

	// Test that the settings page is not accessible as a different user.
	mock.Ctx = handlerutil.WithUser(mock.Ctx, sourcegraph.UserSpec{
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
	mock.Ctx = handlerutil.ClearUser(mock.Ctx)

	if _, err := c.GetOK(router.Rel.URLToUserSubroute(router.UserSettingsProfile, "u").String()); err == nil {
		t.Fatalf("expected to get error when accessing a user's settings page as anonymous. got nil")
	}
}
