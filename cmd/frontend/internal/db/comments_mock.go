package db

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

	"context"
)

type MockComments struct {
	Create          func(ctx context.Context, threadID int32, contents string, authorUserID int32) (*types.Comment, error)
	GetAllForThread func(ctx context.Context, threadID int32) ([]*types.Comment, error)
}

func (s *MockComments) MockCreate(t *testing.T) (called *bool, calledWith *types.Comment) {
	called = new(bool)
	calledWith = &types.Comment{}
	s.Create = func(ctx context.Context, threadID int32, contents string, authorUserID int32) (*types.Comment, error) {
		*called, *calledWith = true, types.Comment{
			ThreadID:     threadID,
			Contents:     contents,
			AuthorUserID: authorUserID,
		}
		return &types.Comment{
			ID:           1,
			ThreadID:     threadID,
			Contents:     contents,
			AuthorUserID: authorUserID,
		}, nil
	}
	return called, calledWith
}
