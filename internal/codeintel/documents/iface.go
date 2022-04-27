package documents

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]Document, error)
}
