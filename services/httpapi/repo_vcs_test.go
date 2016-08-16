package httpapi

import (
	"net/http"
	"reflect"
	"testing"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestRepoResolveRev_ok(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.ResolvedRev{CommitID: "c"}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")

	var res *sourcegraph.ResolvedRev
	if err := c.GetJSON("/repos/r@v/-/rev", &res); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %+v, want %+v", res, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestRepoResolveRev_notFound(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledResolveRev bool
	mock.Repos.ResolveRev_ = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		calledResolveRev = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	resp, err := c.Get("/repos/r@doesntexist/-/rev")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestCommit_ok(t *testing.T) {
	c, mock := newTest()

	want := &vcs.Commit{ID: "c"}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
	calledReposGetCommit := mock.Repos.MockGetCommit_Return_NoCheck(t, want)

	var commit *vcs.Commit
	if err := c.GetJSON("/repos/r@v/-/commit", &commit); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(commit, want) {
		t.Errorf("got %+v, want %+v", commit, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
