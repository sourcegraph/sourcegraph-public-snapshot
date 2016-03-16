package app_test

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestRepo(t *testing.T) {
	c, mock := apptest.New()

	calledResolve := mock.Repos.MockResolve_Local(t, "my/repo")
	calledGet := mockRepoGet(mock, "my/repo")
	calledGetConfig := mockEmptyRepoConfig(mock)
	calledGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledGetReadme := mockNoRepoReadme(mock)
	calledRepoTreeGet := mockEmptyTreeEntry(mock)

	if _, err := c.GetOK(router.Rel.URLToRepo("my/repo").String()); err != nil {
		t.Fatal(err)
	}
	if !*calledResolve {
		t.Error("!calledResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledGetConfig {
		t.Error("!calledGetConfig")
	}
	if !*calledGetCommit {
		t.Error("!calledGetCommit")
	}
	if !*calledGetReadme {
		t.Error("!calledGetReadme")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestRepo_branchWithSlashes(t *testing.T) {
	c, mock := apptest.New()

	mock.Repos.MockResolve_Local(t, "my/repo")
	calledGet := mockRepoGet(mock, "my/repo")
	mockEmptyRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockCurrentSrclibData(mock)
	mockNoRepoReadme(mock)
	mockEmptyTreeEntry(mock)

	url, err := router.Rel.URLToRepoRev("my/repo", "some/branch")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetOK(url.String()); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepo_defaultBranchWithSlashes(t *testing.T) {
	c, mock := apptest.New()

	mock.Repos.MockResolve_Local(t, "my/repo")
	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:           "my/repo",
		DefaultBranch: "some/branch",
	})
	mockEmptyRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockCurrentSrclibData(mock)
	mockNoRepoReadme(mock)
	mockEmptyTreeEntry(mock)

	if _, err := c.GetOK(router.Rel.URLToRepo("my/repo").String()); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

// Test that a "not found" response from the API client causes the
// handler to return a HTTP 404.
func TestRepo_NotFound(t *testing.T) {
	c, mock := apptest.New()

	mock.Repos.MockResolve_Local(t, "my/repo")
	var calledGet bool
	mock.Repos.Get_ = func(context.Context, *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	resp, err := c.Get(router.Rel.URLToRepo("my/repo").String())
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
