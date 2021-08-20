package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

const slowDocumentationPageRequestThreshold = time.Second

// DocumentationPage returns the DocumentationPage for the given PathID.
//
// nil, nil is returned if the page does not exist.
func (r *queryResolver) DocumentationPage(ctx context.Context, pathID string) (_ *precise.DocumentationPageData, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "DocumentationPage", r.operations.documentationPage, slowDocumentationPageRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.String("pathID", pathID),
		},
	})
	defer endObservation()

	for i := range r.uploads {
		traceLog(log.Int("uploadID", r.uploads[i].ID))

		// In the case of multiple LSIF uploads, we merely return the most-recent page from a
		// matching bundle.
		var page *precise.DocumentationPageData
		page, err = r.lsifStore.DocumentationPage(ctx, r.uploads[i].ID, pathID)
		if err == nil {
			return page, nil
		}
	}
	return nil, err
}

const slowDocumentationPathInfoRequestThreshold = time.Second

// DocumentationPathIDInfo returns information about what is located at the given pathID.
//
// nil, nil is returned if the page does not exist.
func (r *queryResolver) DocumentationPathInfo(ctx context.Context, pathID string) (_ *precise.DocumentationPathInfoData, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "DocumentationPathInfo", r.operations.documentationPathInfo, slowDocumentationPathInfoRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.String("pathID", pathID),
		},
	})
	defer endObservation()

	for i := range r.uploads {
		traceLog(log.Int("uploadID", r.uploads[i].ID))

		// In the case of multiple LSIF uploads, we merely return the most-recent info from a
		// matching bundle.
		var pathInfo *precise.DocumentationPathInfoData
		pathInfo, err = r.lsifStore.DocumentationPathInfo(ctx, r.uploads[i].ID, pathID)
		if err == nil {
			return pathInfo, nil
		}
	}
	return nil, err
}

const slowDocumentationRequestThreshold = time.Second

// Documentation returns documentation for the symbol at the given position.
func (r *queryResolver) Documentation(ctx context.Context, line, character int) (_ []*Documentation, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, "Documentation", r.operations.documentation, slowDocumentationRequestThreshold, observation.Args{
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

	// First, perform a definitions request. This handles all the complex logic of finding the
	// symbol, doing cross-repo moniker lookups, etc. for us.
	adjustedLocations, err := r.Definitions(ctx, line, character)
	if err != nil {
		return nil, err
	}
	if len(adjustedLocations) == 0 {
		return nil, nil
	}

	// Now that we have locations resolved to specific dumps, lookup the documentation info.
	documentation := make([]*Documentation, 0, len(adjustedLocations))
	for _, location := range adjustedLocations {
		target := location.AdjustedRange
		pathIDs, err := r.lsifStore.DocumentationAtPosition(
			ctx,
			location.Dump.ID,
			location.Path,
			target.Start.Line,
			target.Start.Character,
		)
		if err != nil {
			return nil, errors.Wrap(err, "lsifStore.DocumentationAtPosition")
		}
		if len(pathIDs) == 0 {
			continue
		}
		for _, pathID := range pathIDs {
			documentation = append(documentation, &Documentation{PathID: pathID})
		}
	}
	return documentation, nil
}
