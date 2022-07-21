package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

const slowStencilRequestThreshold = time.Second

// Stencil return all ranges within a single document.
func (r *queryResolver) Stencil(ctx context.Context) (adjustedRanges []lsifstore.Range, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
	}
	ranges, err := r.symbolsResolver.Stencil(ctx, args)
	for _, r := range ranges {
		adjustedRanges = append(adjustedRanges, sharedRangeTolsifstoreRange(r))
	}
	return adjustedRanges, err
}
