package resolvers

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowHoverRequestThreshold = time.Second

// Hover returns the hover text and range for the symbol at the given position.
func (r *queryResolver) Hover(ctx context.Context, line, character int) (_ string, _ lsifstore.Range, _ bool, err error) {
	ctx, trace, endObservation := observeResolver(ctx, &err, r.operations.hover, slowHoverRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.adjustUploads(ctx, line, character)
	if err != nil {
		return "", lsifstore.Range{}, false, err
	}

	// Keep track of each adjusted range we know about enclosing the requested position.
	//
	// If we don't have hover text within the index where the range is defined, we'll
	// have to look in the definition index and search for the text there. We don't
	// want to return the range associated with the definition, as the range is used
	// as a hint to highlight a range in the current document.
	adjustedRanges := make([]lsifstore.Range, 0, len(adjustedUploads))

	for i := range adjustedUploads {
		adjustedUpload := adjustedUploads[i]
		trace.Log(log.Int("uploadID", adjustedUpload.Upload.ID))

		// Fetch hover text from the index
		text, rn, exists, err := r.lsifStore.Hover(
			ctx,
			adjustedUpload.Upload.ID,
			adjustedUpload.AdjustedPathInBundle,
			adjustedUpload.AdjustedPosition.Line,
			adjustedUpload.AdjustedPosition.Character,
		)
		if err != nil {
			return "", lsifstore.Range{}, false, errors.Wrap(err, "lsifStore.Hover")
		}
		if !exists {
			continue
		}

		// Adjust the highlighted range back to the appropriate range in the target commit
		_, adjustedRange, _, err := r.adjustRange(ctx, r.uploads[i].RepositoryID, r.uploads[i].Commit, r.path, rn)
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if text != "" {
			// Text attached to source range
			return text, adjustedRange, true, nil
		}

		adjustedRanges = append(adjustedRanges, adjustedRange)
	}

	// The Slow path:
	//
	// The indexes we searched in doesn't attach hover text to externally defined symbols.
	// Each indexer is free to make that choice as it's a compromise between ease of development,
	// efficiency of indexing, index output sizes, etc. We can deal with this situation by
	// looking for hover text attached to the precise definition (if one exists).

	// The range we will end up returning is interpreted within the context of the current text
	// document, so any range inside of a remote index would be of no use. We'll return the first
	// (inner-most) range that we adjusted from the source index traversals above.
	var adjustedRange lsifstore.Range
	if len(adjustedRanges) > 0 {
		adjustedRange = adjustedRanges[0]
	}

	// Gather all import monikers attached to the ranges enclosing the requested position
	orderedMonikers, err := r.orderedMonikers(ctx, adjustedUploads, "import")
	if err != nil {
		return "", lsifstore.Range{}, false, err
	}
	trace.Log(
		log.Int("numMonikers", len(orderedMonikers)),
		log.String("monikers", monikersToString(orderedMonikers)),
	)

	// Determine the set of uploads over which we need to perform a moniker search. This will
	// include all all indexes which define one of the ordered monikers. This should not include
	// any of the indexes we have already performed an LSIF graph traversal in above.
	uploads, err := r.definitionUploads(ctx, orderedMonikers)
	if err != nil {
		return "", lsifstore.Range{}, false, err
	}
	trace.Log(
		log.Int("numDefinitionUploads", len(uploads)),
		log.String("definitionUploads", uploadIDsToString(uploads)),
	)

	// Perform the moniker search. This returns a set of locations defining one of the monikers
	// attached to one of the source ranges.
	locations, _, err := r.monikerLocations(ctx, uploads, orderedMonikers, "definitions", DefinitionsLimit, 0)
	if err != nil {
		return "", lsifstore.Range{}, false, err
	}
	trace.Log(log.Int("numLocations", len(locations)))

	for i := range locations {
		// Fetch hover text attached to a definition in the defining index
		text, _, exists, err := r.lsifStore.Hover(
			ctx,
			locations[i].DumpID,
			locations[i].Path,
			locations[i].Range.Start.Line,
			locations[i].Range.Start.Character,
		)
		if err != nil {
			return "", lsifstore.Range{}, false, errors.Wrap(err, "lsifStore.Hover")
		}
		if exists && text != "" {
			// Text attached to definition
			return text, adjustedRange, true, nil
		}
	}

	// No text available
	return "", lsifstore.Range{}, false, nil
}
