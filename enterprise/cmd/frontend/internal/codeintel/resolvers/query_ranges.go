package resolvers

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowRangesRequestThreshold = time.Second

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *queryResolver) Ranges(ctx context.Context, startLine, endLine int) (adjustedRanges []AdjustedCodeIntelligenceRange, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Ranges", r.operations.ranges, slowRangesRequestThreshold, observation.Args{
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
		traceLog(log.Int("uploadID", adjustedUploads[i].Upload.ID))

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
			adjustedRange, err := r.adjustCodeIntelligenceRange(ctx, adjustedUploads[i], rn)
			if err != nil {
				return nil, err
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	traceLog(log.Int("numRanges", len(adjustedRanges)))

	return adjustedRanges, nil
}

// adjustCodeIntelligenceRange translates a range summary (relative to the indexed commit) into an
// equivalent range summary in the requested commit.
func (r *queryResolver) adjustCodeIntelligenceRange(ctx context.Context, upload adjustedUpload, rn lsifstore.CodeIntelligenceRange) (AdjustedCodeIntelligenceRange, error) {
	_, adjustedRange, err := r.adjustRange(ctx, upload.Upload.RepositoryID, upload.Upload.Commit, upload.AdjustedPath, rn.Range)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, err
	}

	uploadsByID := map[int]store.Dump{
		upload.Upload.ID: upload.Upload,
	}

	adjustedDefinitions, err := r.adjustLocations(ctx, uploadsByID, rn.Definitions)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, err
	}

	adjustedReferences, err := r.adjustLocations(ctx, uploadsByID, rn.References)
	if err != nil {
		return AdjustedCodeIntelligenceRange{}, err
	}

	return AdjustedCodeIntelligenceRange{
		Range:       adjustedRange,
		Definitions: adjustedDefinitions,
		References:  adjustedReferences,
		HoverText:   rn.HoverText,
	}, nil
}
