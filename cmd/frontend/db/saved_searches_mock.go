package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSavedSearches struct {
	ListAll                   func(ctx context.Context) ([]api.SavedQuerySpecAndConfig, error)
	ListSavedSearchesByUserID func(ctx context.Context, userID int32) ([]*types.SavedSearch, error)
	Create                    func(ctx context.Context, newSavedSearch *types.SavedSearch) (*types.SavedSearch, error)
	Update                    func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error)
	Delete                    func(ctx context.Context, id int32) error
	GetByID                   func(ctx context.Context, id int32) (*api.SavedQuerySpecAndConfig, error)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_82(size int) error {
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
