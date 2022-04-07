package documents

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/store"
)

type Store interface {
	Todo(ctx context.Context) ([]store.Todo, error)
}
