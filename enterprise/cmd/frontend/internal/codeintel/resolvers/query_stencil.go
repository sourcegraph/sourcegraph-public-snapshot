package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowStencilRequestThreshold = time.Second

// TODO - test
func (r *queryResolver) Stencil(ctx context.Context) (adjustedRanges []lsifstore.Range, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Stencil", r.operations.stencil, slowStencilRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.adjustUploadPaths(ctx)
	if err != nil {
		return nil, err
	}

	for i := range adjustedUploads {
		traceLog(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		// TODO
		ranges, err := r.lsifStore.Stencil(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Stencil")
		}

		for _, rn := range ranges {
			// TODO - adjust range
			// TODO - keep sorted
			adjustedRanges = append(adjustedRanges, rn)
		}
	}
	traceLog(log.Int("numRanges", len(adjustedRanges)))

	return adjustedRanges, nil
}
