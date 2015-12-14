package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Tests that a root commit (i.e., a commit with no parents) can be displayed.
func TestRepoCommit_root(t *testing.T) {
	c, mock := apptest.New()

	mockRepoGet(mock, "my/repo")
	mockCurrentSrclibData(mock)
	mockEnabledRepoConfig(mock)
	mock.Repos.MockGetCommit_Return_NoCheck(t, &vcs.Commit{
		ID:      commitID("a"),
		Parents: nil, // be explicit about this commit being a root commit
	})

	if _, err := c.GetOK(router.Rel.URLToRepoCommit("my/repo", "c").String()); err != nil {
		t.Fatal(err)
	}
}

func TestRepoCommit_general(t *testing.T) {
	c, mock := apptest.New()

	var calledDeltasGet bool
	mockRepoGet(mock, "my/repo")
	mockCurrentSrclibData(mock)
	mockEnabledRepoConfig(mock)
	mock.Repos.MockGetCommit_Return_NoCheck(t, &vcs.Commit{
		ID:      commitID("a"),
		Parents: []vcs.CommitID{commitID("b")},
	})
	mock.Deltas.Get_ = func(ctx context.Context, delta *sourcegraph.DeltaSpec) (*sourcegraph.Delta, error) {
		calledDeltasGet = true
		return &sourcegraph.Delta{}, nil
	}
	mock.Deltas.ListFiles_ = func(ctx context.Context, op *sourcegraph.DeltasListFilesOp) (*sourcegraph.DeltaFiles, error) {
		return &sourcegraph.DeltaFiles{}, nil
	}

	if _, err := c.GetOK(router.Rel.URLToRepoCommit("my/repo", "c").String()); err != nil {
		t.Fatal(err)
	}

	if !calledDeltasGet {
		t.Error("!calledDeltasGet")
	}
}
