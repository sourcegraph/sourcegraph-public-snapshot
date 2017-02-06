package localstore

import (
	"context"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

// repoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type repoVCS struct{}

func (s *repoVCS) Open(ctx context.Context, repo int32) (vcs.Repository, error) {
	if Mocks.RepoVCS.Open != nil {
		return Mocks.RepoVCS.Open(ctx, repo)
	}

	// SECURITY: repository permissions are checked here
	r, err := Repos.Get(ctx, repo)
	if err != nil {
		return nil, err
	}

	return gitcmd.Open(r), nil
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
