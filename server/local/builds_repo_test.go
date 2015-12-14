package local

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstesting "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestBuildsService_GetRepoBuild(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r"}
	want := &sourcegraph.Build{Attempt: 123, Repo: "r", Success: true, CommitID: "c"}

	var calledVCSRepoResolveRevision, calledBuildsGetFirstInCommitOrder bool
	mock.stores.RepoVCS.MockOpen(t, "r", vcstesting.MockRepository{
		ResolveRevision_: func(rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return "c", nil
		},
	})
	mock.stores.Builds.GetFirstInCommitOrder_ = func(context.Context, string, []string, bool) (*sourcegraph.Build, int, error) {
		calledBuildsGetFirstInCommitOrder = true
		return want, 0, nil
	}

	build, err := s.GetRepoBuild(ctx, &testRepoRevSpec)
	if err != nil {
		t.Fatal(err)
	}
	if !calledBuildsGetFirstInCommitOrder {
		t.Error("!calledBuildsGetFirstInCommitOrder")
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("got %+v, want %+v", build, want)
	}
}
