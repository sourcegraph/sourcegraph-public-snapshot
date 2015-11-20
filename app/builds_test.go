package app_test

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestBuilds(t *testing.T) {
	c, mock := apptest.New()

	calledList := mock.Builds.MockList(t, &sourcegraph.Build{Attempt: 1, CommitID: "ASD", Repo: "my/repo"})

	if _, err := c.GetOK(router.Rel.URLTo(router.Builds).String()); err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
}
