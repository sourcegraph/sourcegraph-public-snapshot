package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]Symbol, error)
}
