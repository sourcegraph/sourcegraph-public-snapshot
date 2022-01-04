package database

import (
	"context"
)

type MockUsers struct {
	Count func(ctx context.Context, opt *UsersListOptions) (int, error)
}
