package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockUsers struct {
	GetByID              func(ctx context.Context, id int32) (*types.User, error)
	GetByUsername        func(ctx context.Context, username string) (*types.User, error)
	GetByCurrentAuthUser func(ctx context.Context) (*types.User, error)
	Count                func(ctx context.Context, opt *UsersListOptions) (int, error)
}
