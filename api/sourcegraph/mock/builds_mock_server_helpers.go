package mock

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *BuildsServer) MockGet_Return(t *testing.T, want *sourcegraph.Build) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, op *sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
		*called = true
		return want, nil
	}
	return
}

func (s *BuildsServer) MockList(t *testing.T, want ...*sourcegraph.Build) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, op *sourcegraph.BuildListOptions) (*sourcegraph.BuildList, error) {
		*called = true
		return &sourcegraph.BuildList{Builds: want}, nil
	}
	return
}

func (s *BuildsServer) MockListBuildTasks(t *testing.T, want ...*sourcegraph.BuildTask) (called *bool) {
	called = new(bool)
	s.ListBuildTasks_ = func(ctx context.Context, op *sourcegraph.BuildsListBuildTasksOp) (*sourcegraph.BuildTaskList, error) {
		*called = true
		return &sourcegraph.BuildTaskList{BuildTasks: want}, nil
	}
	return
}
