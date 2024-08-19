package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *gitBlobLSIFDataResolver) Stencil(ctx context.Context) (_ []resolverstubs.RangeResolver, err error) {
	args := codenav.PositionalRequestArgs{
		RequestArgs: codenav.RequestArgs{
			RepositoryID: r.requestState.RepositoryID,
			Commit:       r.requestState.Commit,
		},
		Path: r.requestState.Path,
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.stencil, time.Second, getObservationArgs(&args))
	defer endObservation()

	ranges, err := r.codeNavSvc.GetStencil(ctx, args, r.requestState)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetStencil")
	}

	resolvers := make([]resolverstubs.RangeResolver, 0, len(ranges))
	for _, r := range ranges {
		resolvers = append(resolvers, newRangeResolver(r.ToSCIPRange()))
	}

	return resolvers, nil
}
