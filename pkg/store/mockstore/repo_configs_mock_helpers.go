package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *RepoConfigs) MockGet_Return(t *testing.T, wantRepo int32, returns *sourcegraph.RepoConfig) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo int32) (*sourcegraph.RepoConfig, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "config for repo %v not found", repo)
		}
		return returns, nil
	}
	return
}
