package httpapi

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestBuild(t *testing.T) {
	c, mock := newTest()

	wantBuild := &sourcegraph.Build{CommitID: "c", ID: 123, Repo: 1}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Builds.MockGet_Return(t, wantBuild)

	var build *sourcegraph.Build
	if err := c.GetJSON("/repos/r/-/builds/123", &build); err != nil {
		t.Logf("%#v", build)
		t.Fatal(err)
	}
	if !reflect.DeepEqual(build, wantBuild) {
		t.Errorf("got %+v, want %+v", build, wantBuild)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestBuilds(t *testing.T) {
	c, mock := newTest()

	wantBuilds := &sourcegraph.BuildList{Builds: []*sourcegraph.Build{{ID: 123, CommitID: "c", Repo: 456}}}

	calledReposGet := mock.Repos.MockGet(t, 456)
	calledList := mock.Builds.MockList(t, wantBuilds.Builds...)

	var builds *sourcegraph.BuildList
	if err := c.GetJSON("/builds", &builds); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(builds, wantBuilds) {
		t.Errorf("got %+v, want %+v", builds, wantBuilds)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledList {
		t.Error("!calledList")
	}
}

func TestBuildTasks(t *testing.T) {
	c, mock := newTest()

	wantTasks := &sourcegraph.BuildTaskList{BuildTasks: []*sourcegraph.BuildTask{{ID: 123}}}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledListBuildTasks := mock.Builds.MockListBuildTasks(t, wantTasks.BuildTasks...)

	var tasks *sourcegraph.BuildTaskList
	if err := c.GetJSON("/repos/r/-/builds/123/tasks", &tasks); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tasks, wantTasks) {
		t.Errorf("got %+v, want %+v", tasks, wantTasks)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledListBuildTasks {
		t.Error("!calledListBuildTasks")
	}
}
