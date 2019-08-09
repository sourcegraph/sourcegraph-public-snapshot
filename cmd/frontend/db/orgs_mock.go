package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockOrgs struct {
	GetByID   func(ctx context.Context, id int32) (*types.Org, error)
	GetByName func(ctx context.Context, name string) (*types.Org, error)
	Count     func(ctx context.Context, opt OrgsListOptions) (int, error)
	List      func(ctx context.Context, opt *OrgsListOptions) ([]*types.Org, error)
}

func (s *MockOrgs) MockGetByID_Return(t *testing.T, returns *types.Org, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*types.Org, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_66(size int) error {
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
