package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *Builds) mustCreate(ctx context.Context, t *testing.T, b *sourcegraph.Build) *sourcegraph.Build {
	createdBuild, err := s.Create(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	return createdBuild
}

func (s *Builds) mustCreateBuilds(ctx context.Context, t *testing.T, builds []*sourcegraph.Build) {
	for _, b := range builds {
		s.mustCreate(ctx, t, b)
	}
}

func (s *Builds) mustCreateTasks(ctx context.Context, t *testing.T, tasks []*sourcegraph.BuildTask) []*sourcegraph.BuildTask {
	createdTasks, err := s.CreateTasks(ctx, tasks)
	if err != nil {
		t.Fatal(err)
	}
	return createdTasks
}
