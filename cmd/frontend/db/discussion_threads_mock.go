package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockDiscussionThreads struct {
	Get    func(int64) (*types.DiscussionThread, error)
	Create func(ctx context.Context, newThread *types.DiscussionThread) (*types.DiscussionThread, error)
	Update func(ctx context.Context, threadID int64, opts *DiscussionThreadsUpdateOptions) (*types.DiscussionThread, error)
	List   func(ctx context.Context, opt *DiscussionThreadsListOptions) ([]*types.DiscussionThread, error)
	Count  func(ctx context.Context, opt *DiscussionThreadsListOptions) (int, error)
}

func (s *MockDiscussionThreads) MockCreate_Return(t *testing.T, returns *types.DiscussionThread, returnsErr error) (called *bool, calledWith *types.DiscussionThread) {
	called, calledWith = new(bool), &types.DiscussionThread{}
	s.Create = func(ctx context.Context, newThread *types.DiscussionThread) (*types.DiscussionThread, error) {
		*called = true
		return returns, returnsErr
	}
	return called, calledWith
}

func (s *MockDiscussionThreads) MockUpdate_Return(t *testing.T, returns *types.DiscussionThread, returnsErr error) (called *bool) {
	called = new(bool)
	s.Update = func(ctx context.Context, threadID int64, opts *DiscussionThreadsUpdateOptions) (*types.DiscussionThread, error) {
		*called = true
		return returns, returnsErr
	}
	return called
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_47(size int) error {
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
