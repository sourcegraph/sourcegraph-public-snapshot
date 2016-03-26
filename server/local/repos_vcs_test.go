package local

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func Test_Repos_ListCommits(t *testing.T) {
	rr1 := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "master",
	}
	wantCommits := []*vcs.Commit{
		{ID: vcs.CommitID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")},
		{ID: vcs.CommitID("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")},
		{ID: vcs.CommitID("cccccccccccccccccccccccccccccccccccccccc")},
		{ID: vcs.CommitID("dddddddddddddddddddddddddddddddddddddddd")},
	}

	var s repos
	ctx, mock := testContext()

	mockRepo := vcstest.MockRepository{}
	mockRepo.ResolveRevision_ = func(spec string) (vcs.CommitID, error) {
		if spec != "master" {
			t.Fatalf("call to ResolveRevision with unexpected argument spec=%s", spec)
		}
		return wantCommits[0].ID, nil
	}
	mockRepo.Commits_ = func(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
		if !(opt.Head == wantCommits[0].ID && opt.Base == "") {
			t.Fatalf("call to Commits with unexpected argument opt=%+v", opt)
		}
		return wantCommits, uint(len(wantCommits)), nil
	}
	mock.stores.RepoVCS.Open_ = func(ctx context.Context, repo string) (vcs.Repository, error) {
		return mockRepo, nil
	}

	commitList, err := s.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: rr1.RepoSpec,
		Opt:  &sourcegraph.RepoListCommitsOptions{Head: rr1.Rev},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantCommits, commitList.Commits) {
		t.Errorf("want %+v, got %+v", wantCommits, commitList.Commits)
	}
}
