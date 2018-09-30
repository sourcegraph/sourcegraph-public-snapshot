package db

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"

	"context"
)

type MockDiscussionComments struct {
	Create func(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionComment, error)
	Update func(ctx context.Context, commentID int64, opts *DiscussionCommentsUpdateOptions) (*types.DiscussionComment, error)
	List   func(ctx context.Context, opts *DiscussionCommentsListOptions) ([]*types.DiscussionComment, error)
	Count  func(ctx context.Context, opts *DiscussionCommentsListOptions) (int, error)
}

func (s *MockDiscussionComments) MockCreate(t *testing.T) (called *bool, calledWith *types.DiscussionComment) {
	called = new(bool)
	calledWith = &types.DiscussionComment{}
	s.Create = func(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionComment, error) {
		*called, *calledWith = true, *newComment
		return newComment, nil
	}
	return called, calledWith
}
