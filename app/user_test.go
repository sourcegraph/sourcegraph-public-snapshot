package app_test

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func TestUser(t *testing.T) {
	c, mock := apptest.New()

	calledGet := mock.Users.MockGet(t, "u")
	calledReposList := mock.Repos.MockList(t, "r/r")

	if _, err := c.GetOK(router.Rel.URLToUser("u").String()); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposList {
		t.Error("!calledReposList")
	}
}

func TestUser_NotFound(t *testing.T) {
	c, mock := apptest.New()

	var calledGet bool
	mock.Users.Get_ = func(context.Context, *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		calledGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	resp, err := c.Get(router.Rel.URLToUser("u").String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}
	if !calledGet {
		t.Error("!calledGet")
	}
}

func TestUser_Disabled(t *testing.T) {
	c, mock := apptest.New()

	calledGet := mock.Users.MockGet_Return(t, &sourcegraph.User{
		Login:    "u",
		Disabled: true,
	})

	resp, err := c.Get(router.Rel.URLToUser("u").String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}
