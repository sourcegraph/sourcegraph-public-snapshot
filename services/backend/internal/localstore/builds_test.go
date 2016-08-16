package localstore

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *builds) mustCreate(ctx context.Context, t *testing.T, b *sourcegraph.Build) *sourcegraph.Build {
	createdBuild, err := s.Create(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	return createdBuild
}

func (s *builds) mustCreateBuilds(ctx context.Context, t *testing.T, builds []*sourcegraph.Build) {
	for _, b := range builds {
		s.mustCreate(ctx, t, b)
	}
}

func (s *builds) mustCreateTasks(ctx context.Context, t *testing.T, tasks []*sourcegraph.BuildTask) {
	_, err := s.CreateTasks(ctx, tasks)
	if err != nil {
		t.Fatal(err)
	}
}
