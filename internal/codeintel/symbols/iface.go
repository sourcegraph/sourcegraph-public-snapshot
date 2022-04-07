package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/store"
)

type Store interface {
	Todo(ctx context.Context) ([]store.Todo, error)
}
