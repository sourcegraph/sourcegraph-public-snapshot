package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockComments struct {
	Create func(ctx context.Context, threadID int32, contents string, authorName, authorEmail string) (*sourcegraph.Comment, error)
}

func (s *MockComments) MockCreate(t *testing.T) (called *bool, calledWith *sourcegraph.Comment) {
	called = new(bool)
	calledWith = &sourcegraph.Comment{}
	s.Create = func(ctx context.Context, threadID int32, contents string, authorName, authorEmail string) (*sourcegraph.Comment, error) {
		*called, *calledWith = true, sourcegraph.Comment{
			ThreadID:    threadID,
			Contents:    contents,
			AuthorName:  authorName,
			AuthorEmail: authorEmail,
		}
		return &sourcegraph.Comment{
			ID:          1,
			ThreadID:    threadID,
			Contents:    contents,
			AuthorName:  authorName,
			AuthorEmail: authorEmail,
		}, nil
	}
	return called, calledWith
}
