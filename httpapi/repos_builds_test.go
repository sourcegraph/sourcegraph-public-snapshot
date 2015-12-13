package httpapi

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestRepoBuild(t *testing.T) {
	c, mock := newTest()

	wantBuild := &sourcegraph.Build{Host: "abc"}

	calledRepoGet := mock.Repos.MockGet(t, "r/r")
	calledRepoGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledGetRepoBuild := mock.Builds.MockGetRepoBuild(t, wantBuild)

	var build *sourcegraph.Build
	if err := c.GetJSON("/repos/r/r@c/.build", &build); err != nil {
		t.Fatal(err)
	}
	if !*calledRepoGet {
		t.Error("!calledRepoGet")
	}
	if !*calledRepoGetCommit {
		t.Error("!calledRepoGetCommit")
	}
	if !*calledGetRepoBuild {
		t.Error("!calledGetRepoBuild")
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
}

func TestRepoBuildsCreate(t *testing.T) {
	c, mock := newTest()

	wantBuild := &sourcegraph.Build{ID: 123, Repo: "r/r", CommitID: "c"}

	calledRepoGet := mock.Repos.MockGet(t, "r/r")
	calledRepoGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	var calledCreate bool
	mock.Builds.Create_ = func(ctx context.Context, op *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		calledCreate = true
		return wantBuild, nil
	}

	var build *sourcegraph.Build
	if err := c.DoJSON("POST", "/repos/r/r@c/.builds", &sourcegraph.BuildCreateOptions{}, &build); err != nil {
		t.Fatal(err)
	}
	if !*calledRepoGet {
		t.Error("!calledRepoGet")
	}
	if !*calledRepoGetCommit {
		t.Error("!calledRepoGetCommit")
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
}
