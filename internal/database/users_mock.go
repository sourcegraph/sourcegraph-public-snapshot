package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockUsers struct {
	GetByCurrentAuthUser func(ctx context.Context) (*types.User, error)
	Count                func(ctx context.Context, opt *UsersListOptions) (int, error)
}
