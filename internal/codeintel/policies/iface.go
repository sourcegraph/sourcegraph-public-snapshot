package policies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
)

type Store interface {
	List(ctx context.Context, opts store.ListOpts) ([]Policy, error)
}
