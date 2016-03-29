package httpapi

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestRepoBuildsCreate(t *testing.T) {
	c, mock := newTest()

	wantBuild := &sourcegraph.Build{ID: 123, Repo: "r/r", CommitID: "c"}

	calledRepoGet := mock.Repos.MockGet(t, "r/r")
	var calledCreate bool
	mock.Builds.Create_ = func(ctx context.Context, op *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		calledCreate = true
		return wantBuild, nil
	}

	var build *sourcegraph.Build
	if err := c.DoJSON("POST", "/repos/r/r/-/builds", &sourcegraph.BuildsCreateOp{}, &build); err != nil {
		t.Fatal(err)
	}
	if !*calledRepoGet {
		t.Error("!calledRepoGet")
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
}
