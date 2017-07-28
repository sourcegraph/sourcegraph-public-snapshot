package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockComments struct {
	Create func(ctx context.Context, newComment *sourcegraph.Comment) (*sourcegraph.Comment, error)
}

func (s *MockComments) MockCreate(t *testing.T) (called *bool, calledWith *sourcegraph.Comment) {
	called = new(bool)
	calledWith = &sourcegraph.Comment{}
	s.Create = func(ctx context.Context, newComment *sourcegraph.Comment) (*sourcegraph.Comment, error) {
		*called, *calledWith = true, *newComment
		return newComment, nil
	}
	return called, calledWith
}
