package mockstore

import (
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sqs/pbtypes"
)

func (s *Builds) MockGet(t *testing.T, wantBuild sourcegraph.BuildSpec) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, build sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
		*called = true
		if build != wantBuild {
			t.Errorf("got build %q, want %q", build, wantBuild)
			return nil, sourcegraph.ErrBuildNotFound
		}
		return &sourcegraph.Build{Attempt: build.Attempt, CommitID: build.CommitID, Repo: build.Repo.URI}, nil
	}
	return
}

func (s *Builds) MockGet_Return(t *testing.T, returns *sourcegraph.Build) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, build sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
		*called = true
		if build != returns.Spec() {
			t.Errorf("got build %q, want %q", build, returns.Spec())
			return nil, sourcegraph.ErrBuildNotFound
		}
		return returns, nil
	}
	return
}

func (s *Builds) MockList(t *testing.T, wantBuilds ...sourcegraph.BuildSpec) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
		*called = true
		builds := make([]*sourcegraph.Build, len(wantBuilds))
		for i, build := range wantBuilds {
			builds[i] = &sourcegraph.Build{Attempt: build.Attempt, CommitID: build.CommitID, Repo: build.Repo.URI, CreatedAt: pbtypes.NewTimestamp(time.Unix(int64(len(wantBuilds)-1-i), 0))}
		}
		builds = store.SortAndPaginateBuilds(builds, opt)
		return builds, nil
	}
	return
}
