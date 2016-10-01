package mockstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type GlobalRefs struct {
	Get    func(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error)
	Update func(ctx context.Context, op store.RefreshIndexOp) error
}

type GlobalDeps struct {
	Upsert  func(ctx context.Context, resolutions []*unit.Resolution) error
	Resolve func(ctx context.Context, raw *unit.Key) ([]unit.Key, error)
}

type Defs struct {
	Search func(ctx context.Context, op store.DefSearchOp) (*sourcegraph.SearchResultsList, error)
	Update func(ctx context.Context, op store.RefreshIndexOp) error
}
