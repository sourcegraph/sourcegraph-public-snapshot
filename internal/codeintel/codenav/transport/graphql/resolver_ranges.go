package graphql

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowRangesRequestThreshold = time.Second

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *resolver) Ranges(ctx context.Context, args shared.RequestArgs, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.ranges, slowRangesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("numUploads", len(r.dataLoader.uploads)),
			log.String("uploads", uploadIDsToString(r.dataLoader.uploads)),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer endObservation()

	uploadsWithPath, err := r.getUploadPaths(ctx, args.Path)
	if err != nil {
		return nil, err
	}

	for i := range uploadsWithPath {
		trace.Log(log.Int("uploadID", uploadsWithPath[i].Upload.ID))

		ranges, err := r.svc.GetRanges(
			ctx,
			uploadsWithPath[i].Upload.ID,
			uploadsWithPath[i].TargetPathWithoutRoot,
			startLine,
			endLine,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Ranges")
		}

		for _, rn := range ranges {
			adjustedRange, ok, err := r.getCodeIntelligenceRange(ctx, uploadsWithPath[i], rn)
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

// getCodeIntelligenceRange translates a range summary (relative to the indexed commit) into an
// equivalent range summary in the requested commit. If the translation fails, a false-valued flag
// is returned.
func (r *resolver) getCodeIntelligenceRange(ctx context.Context, upload visibleUpload, rn shared.CodeIntelligenceRange) (shared.AdjustedCodeIntelligenceRange, bool, error) {
	_, adjustedRange, ok, err := r.getSourceRange(ctx, upload.Upload.RepositoryID, upload.Upload.Commit, upload.TargetPath, rn.Range)
	if err != nil || !ok {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	definitions, err := r.getUploadLocations(ctx, rn.Definitions)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	references, err := r.getUploadLocations(ctx, rn.References)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	implementations, err := r.getUploadLocations(ctx, rn.Implementations)
	if err != nil {
		return shared.AdjustedCodeIntelligenceRange{}, false, err
	}

	return shared.AdjustedCodeIntelligenceRange{
		Range:           adjustedRange,
		Definitions:     definitions,
		References:      references,
		Implementations: implementations,
		HoverText:       rn.HoverText,
	}, true, nil
}
