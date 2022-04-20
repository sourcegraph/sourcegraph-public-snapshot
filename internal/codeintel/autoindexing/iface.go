package autoindexing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]IndexJob, error)
}
