package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func TestRepoSearch(t *testing.T) {
	c, mock := apptest.New()

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Builds.GetRepoBuildInfo_ = func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		return &sourcegraph.RepoBuildInfo{
			LastSuccessful: &sourcegraph.Build{CommitID: "c2"},
		}, nil
	}
	mockEnabledRepoConfig(mock)

	u, err := router.Rel.URLToRepoSearch("my/repo", "", "myquery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetOK(u.String()); err != nil {
		t.Fatal(err)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
