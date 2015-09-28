package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstesting "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
)

func (s *RepoVCS) MockOpen(t *testing.T, wantRepo string, mockVCSRepo vcstesting.MockRepository) (called *bool) {
	called = new(bool)
	s.Open_ = func(ctx context.Context, repo string) (vcs.Repository, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %q, want %q", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return mockVCSRepo, nil
	}
	return
}
