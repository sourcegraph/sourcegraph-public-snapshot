package mockstore

import (
	"testing"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func (s *RepoVCS) MockOpen(t *testing.T, wantRepo int32, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open_ = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return mockVCSRepo, nil
	}
	return
}

func (s *RepoVCS) MockOpen_NoCheck(t *testing.T, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open_ = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		*called = true
		return mockVCSRepo, nil
	}
	return
}
