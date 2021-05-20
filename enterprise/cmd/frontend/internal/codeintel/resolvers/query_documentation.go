package resolvers

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

const slowDocumentationPageRequestThreshold = time.Second

// DocumentationPage returns the DocumentationPage for the given PathID.
//
// nil, nil is returned if the page does not exist.
func (r *queryResolver) DocumentationPage(ctx context.Context, pathID string) (_ *semantic.DocumentationPageData, err error) {
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
		var page *semantic.DocumentationPageData
		page, err = r.lsifStore.DocumentationPage(ctx, r.uploads[i].ID, pathID)
		if err == nil {
			return page, nil
		}
	}

	return nil, err
}
