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

func TestRepo_unbuilt(t *testing.T) {
	c, mock := apptest.New()

	calledGet := mockRepoGet(mock, "my/repo")
	calledGetConfig := mockEnabledRepoConfig(mock)
	calledGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledBuildsGetRepoBuildInfo := mockNoRepoBuild(mock)

	if _, err := c.GetOK(router.Rel.URLToRepo("my/repo").String()); err != nil {
		t.Fatal(err)
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
	if !*calledBuildsGetRepoBuildInfo {
		t.Error("!calledBuildsGetRepoBuildInfo")
	}
}

func TestRepo_built(t *testing.T) {
	c, mock := apptest.New()

	calledGet := mockRepoGet(mock, "my/repo")
	calledGetConfig := mockEnabledRepoConfig(mock)
	calledGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledBuildsGetRepoBuildInfo := mockCurrentRepoBuild(mock)
	calledGetReadme := mockNoRepoReadme(mock)
	calledRepoTreeGet := mockEmptyTreeEntry(mock)

	if _, err := c.GetOK(router.Rel.URLToRepo("my/repo").String()); err != nil {
		t.Fatal(err)
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
	if !*calledBuildsGetRepoBuildInfo {
		t.Error("!calledBuildsGetRepoBuildInfo")
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

	calledGet := mockRepoGet(mock, "my/repo")
	mockEnabledRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockCurrentRepoBuild(mock)
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

	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:           "my/repo",
		DefaultBranch: "some/branch",
	})
	mockEnabledRepoConfig(mock)
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockCurrentRepoBuild(mock)
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

func TestRepo_GitHubProfileRedirect(t *testing.T) {
	c, _ := apptest.New()

	resp, err := c.GetNoFollowRedirects("/github.com/u")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusMovedPermanently; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}
	if got, want := resp.Header.Get("location"), "/~u"; got != want {
		t.Errorf("got Location == %q, want %q", got, want)
	}
}
