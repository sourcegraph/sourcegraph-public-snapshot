package db

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type MockSharedItems struct {
	Create func(ctx context.Context, item *types.SharedItem) (*url.URL, error)
	Get    func(ctx context.Context, ulid string) (*types.SharedItem, error)
}

func (s *MockSharedItems) MockCreate(t *testing.T) (called *bool, calledWith *types.SharedItem) {
	called = new(bool)
	calledWith = new(types.SharedItem)
	s.Create = func(ctx context.Context, item *types.SharedItem) (*url.URL, error) {
		*called, *calledWith = true, *item
		return &url.URL{Path: fmt.Sprintf("ulid-thread-%d-comment-%d", item.ThreadID, item.CommentID)}, nil
	}
	return called, calledWith
}

func (s *MockSharedItems) MockGet(t *testing.T, item *types.SharedItem) (called *bool, calledWith *string) {
	called = new(bool)
	calledWith = new(string)
	s.Get = func(ctx context.Context, ulid string) (*types.SharedItem, error) {
		*called, *calledWith = true, ulid
		return item, nil
	}
	return called, calledWith
}
