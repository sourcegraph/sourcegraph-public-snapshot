package localstore

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSharedItems struct {
	Create func(ctx context.Context, item *sourcegraph.SharedItem) (*url.URL, error)
	Get    func(ctx context.Context, ulid string) (*sourcegraph.SharedItem, error)
}

func (s *MockSharedItems) MockCreate(t *testing.T) (called *bool, calledWith *sourcegraph.SharedItem) {
	called = new(bool)
	calledWith = new(sourcegraph.SharedItem)
	s.Create = func(ctx context.Context, item *sourcegraph.SharedItem) (*url.URL, error) {
		*called, *calledWith = true, *item
		return &url.URL{Path: fmt.Sprintf("ulid-thread-%d-comment-%d", item.ThreadID, item.CommentID)}, nil
	}
	return called, calledWith
}

func (s *MockSharedItems) MockGet(t *testing.T, item *sourcegraph.SharedItem) (called *bool, calledWith *string) {
	called = new(bool)
	calledWith = new(string)
	s.Get = func(ctx context.Context, ulid string) (*sourcegraph.SharedItem, error) {
		*called, *calledWith = true, ulid
		return item, nil
	}
	return called, calledWith
}
