package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockDiscussionComments struct {
	Create func(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionComment, error)
	Update func(ctx context.Context, commentID int64, opts *DiscussionCommentsUpdateOptions) (*types.DiscussionComment, error)
	List   func(ctx context.Context, opts *DiscussionCommentsListOptions) ([]*types.DiscussionComment, error)
	Get    func(commentID int64) (*types.DiscussionComment, error)
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_42(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
