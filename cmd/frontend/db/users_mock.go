package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockUsers struct {
	Create               func(ctx context.Context, info NewUser) (newUser *types.User, err error)
	Update               func(userID int32, update UserUpdate) error
	SetIsSiteAdmin       func(id int32, isSiteAdmin bool) error
	GetByID              func(ctx context.Context, id int32) (*types.User, error)
	GetByUsername        func(ctx context.Context, username string) (*types.User, error)
	GetByCurrentAuthUser func(ctx context.Context) (*types.User, error)
	GetByVerifiedEmail   func(ctx context.Context, email string) (*types.User, error)
	Count                func(ctx context.Context, opt *UsersListOptions) (int, error)
	List                 func(ctx context.Context, opt *UsersListOptions) ([]*types.User, error)
}

func (s *MockUsers) MockGetByID_Return(t *testing.T, returns *types.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_100(size int) error {
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
