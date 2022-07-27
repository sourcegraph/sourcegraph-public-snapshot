package graphql

import (
	"context"
	"sort"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowStencilRequestThreshold = time.Second

// Stencil returns all ranges within a single document.
func (r *resolver) Stencil(ctx context.Context, args shared.RequestArgs) (adjustedRanges []shared.Range, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.stencil, slowStencilRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("numUploads", len(r.dataLoader.uploads)),
			log.String("uploads", uploadIDsToString(r.dataLoader.uploads)),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.getUploadPaths(ctx, args.Path)
	if err != nil {
		return nil, err
	}

	for i := range adjustedUploads {
		trace.Log(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		ranges, err := r.svc.GetStencil(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].TargetPathWithoutRoot,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.Stencil")
		}

		for _, rn := range ranges {
			// Adjust the highlighted range back to the appropriate range in the target commit
			_, adjustedRange, _, err := r.getSourceRange(ctx, r.dataLoader.uploads[i].RepositoryID, r.dataLoader.uploads[i].Commit, args.Path, rn)
			if err != nil {
				return nil, err
			}

			adjustedRanges = append(adjustedRanges, adjustedRange)
		}
	}
	trace.Log(log.Int("numRanges", len(adjustedRanges)))

	return sortRanges(adjustedRanges), nil
}

func sortRanges(ranges []shared.Range) []shared.Range {
	sort.Slice(ranges, func(i, j int) bool {
		iStart := ranges[i].Start
		jStart := ranges[j].Start

		if iStart.Line < jStart.Line {
			// iStart comes first
			return true
		} else if iStart.Line > jStart.Line {
			// jStart comes first
			return false
		}
		// otherwise, starts on same line

		if iStart.Character < jStart.Character {
			// iStart comes first
			return true
		} else if iStart.Character > jStart.Character {
			// jStart comes first
			return false
		}
		// otherwise, starts at same character

		iEnd := ranges[i].End
		jEnd := ranges[j].End

		if jEnd.Line < iEnd.Line {
			// ranges[i] encloses ranges[j] (we want smaller first)
			return false
		} else if jStart.Line < jEnd.Line {
			// ranges[j] encloses ranges[i] (we want smaller first)
			return true
		}
		// otherwise, ends on same line

		if jStart.Character < jEnd.Character {
			// ranges[j] encloses ranges[i] (we want smaller first)
			return true
		}

		return false
	})

	return ranges
}
