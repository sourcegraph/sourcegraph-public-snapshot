package httpapi

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestRepoBuildsCreate(t *testing.T) {
	c, mock := newTest()

	wantBuild := &sourcegraph.Build{ID: 123, Repo: 1, CommitID: "c"}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledCreate bool
	mock.Builds.Create_ = func(ctx context.Context, op *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		calledCreate = true
		return wantBuild, nil
	}

	var build *sourcegraph.Build
	if err := c.DoJSON("POST", "/repos/r/-/builds", &sourcegraph.BuildsCreateOp{}, &build); err != nil {
		t.Fatal(err)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
}
