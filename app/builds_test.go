package app_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestBuilds(t *testing.T) {
	c, mock := apptest.New()

	calledList := mock.Builds.MockList(t, &sourcegraph.Build{ID: 1, CommitID: strings.Repeat("a", 40), Repo: "my/repo"})
	calledGetCommit := mockRepoCommit(mock, &vcs.Commit{})

	if _, err := c.GetOK(router.Rel.URLTo(router.Builds).String()); err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !*calledGetCommit {
		t.Error("!calledGetCommit")
	}
}
