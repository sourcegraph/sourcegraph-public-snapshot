package localstore

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// repoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type repoVCS struct{}

// getRepoDir gets the dir (relative to the base repo VCS storage dir)
// where the repo's git repository data lives.
func getRepoDir(ctx context.Context, repo int32) (string, error) {
	dir, err := appDBH(ctx).SelectStr("SELECT uri FROM repo WHERE id=$1;", repo)
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", grpc.Errorf(codes.NotFound, "repo not found (looking up dir): %d", repo)
	}
	return dir, nil
}

func (s *repoVCS) Open(ctx context.Context, repo int32) (vcs.Repository, error) {
	if Mocks.RepoVCS.Open != nil {
		return Mocks.RepoVCS.Open(ctx, repo)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoVCS.Open", repo); err != nil {
		return nil, err
	}
	dir, err := getRepoDir(ctx, repo)
	if err != nil {
		return nil, err
	}

	return gitcmd.Open(dir), nil
}

// CloneInfo is the information needed to clone a repository.
type CloneInfo struct {
	// VCS is the type of VCS (e.g., "git")
	VCS string
	// CloneURL is the remote URL from which to clone.
	CloneURL string
	// Additional options
	vcs.RemoteOpts
}

func (s *repoVCS) Clone(ctx context.Context, repo int32, info *CloneInfo) error {
	if Mocks.RepoVCS.Clone != nil {
		return Mocks.RepoVCS.Clone(ctx, repo, info)
	}

	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "RepoVCS.Clone", repo); err != nil {
		return err
	}
	dir, err := getRepoDir(ctx, repo)
	if err != nil {
		return err
	}

	return gitserver.DefaultClient.Clone(ctx, dir, info.CloneURL, &info.RemoteOpts)
}

type MockRepoVCS struct {
	Open  func(ctx context.Context, repo int32) (vcs.Repository, error)
	Clone func(ctx context.Context, repo int32, info *CloneInfo) error
}

func (s *MockRepoVCS) MockOpen(t *testing.T, wantRepo int32, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return mockVCSRepo, nil
	}
	return
}

func (s *MockRepoVCS) MockOpen_NoCheck(t *testing.T, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		*called = true
		return mockVCSRepo, nil
	}
	return
}
