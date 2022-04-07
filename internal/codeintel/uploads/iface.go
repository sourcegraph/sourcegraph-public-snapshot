package uploads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/store"
)

type Store interface {
	Todo(ctx context.Context) ([]store.Todo, error)
}
