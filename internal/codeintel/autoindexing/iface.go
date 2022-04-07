package autoindexing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/store"
)

type Store interface {
	Todo(ctx context.Context) ([]store.Todo, error)
}
