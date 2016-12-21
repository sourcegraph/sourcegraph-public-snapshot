package localstore

import (
	"context"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
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
		return "", legacyerr.Errorf(legacyerr.NotFound, "repo not found (looking up dir): %d", repo)
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

type MockRepoVCS struct {
	Open func(ctx context.Context, repo int32) (vcs.Repository, error)
}

func (s *MockRepoVCS) MockOpen(t *testing.T, wantRepo int32, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %v not found", wantRepo)
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
