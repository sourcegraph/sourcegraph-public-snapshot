package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *RepoConfigs) MockGet_Return(t *testing.T, wantRepo string, returns *sourcegraph.RepoConfig) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %q, want %q", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "config for repo %v not found", repo)
		}
		return returns, nil
	}
	return
}
