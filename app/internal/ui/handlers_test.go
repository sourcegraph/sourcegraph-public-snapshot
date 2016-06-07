package ui

import (
	"fmt"
	"net/http"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptestutil"
)

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	return apptest.New()
}

func getStatus(c interface {
	Get(url string) (*http.Response, error)
}, url string, wantStatusCode int) error {
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatusCode {
		return fmt.Errorf("got HTTP %d, want %d", resp.StatusCode, wantStatusCode)
	}
	return nil
}

func TestCatchAll(t *testing.T) {
	c, _ := newTest()
	if err := getStatus(c, "/tools", http.StatusOK); err != nil {
		t.Fatal(err)
	}
}

var urls = map[string]struct {
	repo string // repo is necessary (but not sufficient) for this route
	rev  string // rev is necessary (but not sufficient) for this route
}{
	"/r":                  {repo: "r"},
	"/r@v":                {repo: "r", rev: "v"},
	"/r@v/-/tree/d":       {repo: "r", rev: "v"},
	"/r@v/-/blob/f":       {repo: "r", rev: "v"},
	"/r@v/-/def/t/u/-/p":  {repo: "r", rev: "v"},
	"/r@v/-/info/t/u/-/p": {repo: "r", rev: "v"},
	"/r/-/builds":         {repo: "r"},
	"/r/-/builds/2":       {repo: "r"},
}

func TestRepo_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Repos.MockGet(t, 1)
	// (Should not try to resolve the revision; see serveRepo for why.)

	if err := getStatus(c, "/r", http.StatusOK); err != nil {
		t.Fatal(err)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepo_Error_Resolve(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_NotFound(t, req.repo)

		if err := getStatus(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
	}
}

func TestRepo_Error_Get(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		var calledGet bool
		mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
			calledGet = true
			return nil, grpc.Errorf(codes.NotFound, "")
		}

		if err := getStatus(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
		if !calledGet {
			t.Errorf("%s: !calledGet", url)
		}
	}
}

func TestRepoRev_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Repos.MockGet(t, 1)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")

	if err := getStatus(c, "/r@v", http.StatusOK); err != nil {
		t.Fatal(err)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestRepoRev_Error(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := mock.Repos.MockGet(t, 1)
		var calledReposResolveRev bool
		mock.Repos.ResolveRev_ = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
			calledReposResolveRev = true
			return nil, grpc.Errorf(codes.NotFound, "")
		}

		if err := getStatus(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
		if !*calledGet {
			t.Errorf("%s: !calledGet", url)
		}
		if !calledReposResolveRev {
			t.Errorf("%s: !calledReposResolveRev", url)
		}
	}
}
