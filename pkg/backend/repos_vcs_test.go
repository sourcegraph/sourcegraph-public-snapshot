package backend

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func TestReposService_resolveRev_noRevSpecified_getsDefaultBranch(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return vcs.CommitID(want), nil
		},
	})

	// (no rev/branch specified)
	commitID, err := resolveRepoRev(ctx, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestReposService_resolveRev_noCommitIDSpecified_resolvesRev(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return vcs.CommitID(want), nil
		},
	})

	commitID, err := resolveRepoRev(ctx, 1, "b")
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestReposService_resolveRev_commitIDSpecified_resolvesCommitID(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return vcs.CommitID(want), nil
		},
	})

	commitID, err := resolveRepoRev(ctx, 1, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestReposService_resolveRev_commitIDSpecified_failsToResolve(t *testing.T) {
	ctx := testContext()

	want := errors.New("x")

	var calledVCSRepoResolveRevision bool
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return "", errors.New("x")
		},
	})

	_, err := resolveRepoRev(ctx, 1, strings.Repeat("a", 40))
	if !reflect.DeepEqual(err, want) {
		t.Fatalf("got err %v, want %v", err, want)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
}

func Test_Repos_ListCommits(t *testing.T) {
	wantCommits := []*vcs.Commit{
		{ID: vcs.CommitID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")},
		{ID: vcs.CommitID("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")},
		{ID: vcs.CommitID("cccccccccccccccccccccccccccccccccccccccc")},
		{ID: vcs.CommitID("dddddddddddddddddddddddddddddddddddddddd")},
	}

	var s repos
	ctx := testContext()

	calledGet := Mocks.Repos.MockGet(t, 1)
	mockRepo := vcstest.MockRepository{}
	mockRepo.ResolveRevision_ = func(ctx context.Context, spec string) (vcs.CommitID, error) {
		if spec != "v" {
			t.Fatalf("call to ResolveRevision with unexpected argument spec=%s", spec)
		}
		return wantCommits[0].ID, nil
	}
	mockRepo.Commits_ = func(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
		if !(opt.Head == wantCommits[0].ID && opt.Base == "") {
			t.Fatalf("call to Commits with unexpected argument opt=%+v", opt)
		}
		return wantCommits, uint(len(wantCommits)), nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	commitList, err := s.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: 1,
		Opt:  &sourcegraph.RepoListCommitsOptions{Head: "v"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantCommits, commitList.Commits) {
		t.Errorf("want %+v, got %+v", wantCommits, commitList.Commits)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func Test_RepoResolve_DeleteOrKeepRepo(t *testing.T) {
	cases := []struct {
		vcsErr       error
		expectDelete bool
	}{
		{nil, false},
		{vcs.RepoNotExistError{CloneInProgress: true}, false},
		{vcs.RepoNotExistError{CloneInProgress: false}, true},
	}

	for _, testcase := range cases {
		testReposResolveDeleteOrKeepRepo(t, testcase.vcsErr, testcase.expectDelete)
	}
}

func testReposResolveDeleteOrKeepRepo(t *testing.T, vcsErr error, expectDelete bool) {
	var s repos
	ctx := testContext()
	requestedRepoID := int32(42)

	expCalledDelete := map[int32]struct{}{}
	if expectDelete {
		expCalledDelete[requestedRepoID] = struct{}{}
	}

	mockNonExistentRepo := vcstest.MockRepository{}
	mockNonExistentRepo.ResolveRevision_ = func(ctx context.Context, spec string) (vcs.CommitID, error) {
		return "", vcsErr
	}

	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockNonExistentRepo, nil
	}
	calledDelete := map[int32]struct{}{}
	localstore.Mocks.Repos.Delete = func(ctx context.Context, repo int32) error {
		calledDelete[repo] = struct{}{}
		return nil
	}

	// Test ResolveRev
	_, err := s.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: requestedRepoID, Rev: "master"})
	if !reflect.DeepEqual(err, vcsErr) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calledDelete, expCalledDelete) {
		t.Errorf("Expected delete calls to be %+v, actual was %+v", expCalledDelete, calledDelete)
	}
}
