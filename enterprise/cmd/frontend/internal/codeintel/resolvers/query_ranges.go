package resolvers

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowRangesRequestThreshold = time.Second

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *queryResolver) Ranges(ctx context.Context, startLine, endLine int) (adjustedRanges []AdjustedCodeIntelligenceRange, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.ranges, slowRangesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.adjustUploadPaths(ctx)
	if err != nil {
		return nil, err
	}

	for i := range adjustedUploads {
		trace.Log(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		ranges, err := r.lsifStore.Ranges(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			startLine, // TODO - adjust these as well
			endLine,   // TODO - adjust these as well
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Ranges")
		}

		for _, rn := range ranges {
			adjustedRange, ok, err := r.adjustCodeIntelligenceRange(ctx, adjustedUploads[i], rn)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.Log(log.Int("numRanges", len(adjustedRanges)))

	return adjustedRanges, nil
}

// adjustCodeIntelligenceRange translates a range summary (relative to the indexed commit) into an
// equivalent range summary in the requested commit. If the translation fails, a false-valued flag
// is returned.
func (r *queryResolver) adjustCodeIntelligenceRange(ctx context.Context, upload adjustedUpload, rn lsifstore.CodeIntelligenceRange) (AdjustedCodeIntelligenceRange, bool, error) {
	_, adjustedRange, ok, err := r.adjustRange(ctx, upload.Upload.RepositoryID, upload.Upload.Commit, upload.AdjustedPath, rn.Range)
	if err != nil || !ok {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	adjustedDefinitions, err := r.adjustLocations(ctx, rn.Definitions)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	adjustedReferences, err := r.adjustLocations(ctx, rn.References)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	adjustedImplementations, err := r.adjustLocations(ctx, rn.Implementations)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, false, err
	}

	return AdjustedCodeIntelligenceRange{
		Range:           adjustedRange,
		Definitions:     adjustedDefinitions,
		References:      adjustedReferences,
		Implementations: adjustedImplementations,
		HoverText:       rn.HoverText,
	}, true, nil
}
