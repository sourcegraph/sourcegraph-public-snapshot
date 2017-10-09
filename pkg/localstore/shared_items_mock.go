package localstore

import (
	"context"
	"fmt"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSharedItems struct {
	Create func(ctx context.Context, item *sourcegraph.SharedItem) (string, error)
	Get    func(ctx context.Context, ulid string) (*sourcegraph.SharedItem, error)
}

func (s *MockSharedItems) MockCreate(t *testing.T) (called *bool, calledWith *sourcegraph.SharedItem) {
	called = new(bool)
	calledWith = new(sourcegraph.SharedItem)
	s.Create = func(ctx context.Context, item *sourcegraph.SharedItem) (string, error) {
		*called, *calledWith = true, *item
		switch {
		case item.ThreadID != nil:
			return fmt.Sprintf("ulid-thread-%d", item.ThreadID), nil
		case item.CommentID != nil:
			return fmt.Sprintf("ulid-comment-%d", item.CommentID), nil
		default:
			panic("never here")
		}
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
