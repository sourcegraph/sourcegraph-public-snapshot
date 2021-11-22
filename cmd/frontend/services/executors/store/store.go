package store

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ExecutorStore interface {
	List(ctx context.Context, args ExecutorStoreListOptions) ([]types.Executor, int, error)
	GetByID(ctx context.Context, id int) (types.Executor, bool, error)
	Heartbeat(ctx context.Context, executor types.Executor) error
}

type ExecutorStoreListOptions struct {
	Query  string
	Active bool
	Offset int
	Limit  int
}
