package uploads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]Upload, error)
}
