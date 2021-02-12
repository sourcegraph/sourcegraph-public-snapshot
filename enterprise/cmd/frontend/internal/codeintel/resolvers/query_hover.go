package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowHoverRequestThreshold = time.Second

// Hover returns the hover text and range for the symbol at the given position.
func (r *queryResolver) Hover(ctx context.Context, line, character int) (_ string, _ lsifstore.Range, _ bool, err error) {
	ctx, endObservation := observeResolver(ctx, &err, "Hover", r.operations.hover, slowHoverRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.String("uploadIDs", strings.Join(r.uploadIDs(), ", ")),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation()

	for i := range r.uploads {
		text, rn, exists, err := r.hoverForUpload(ctx, line, character, r.uploads[i])
		if err != nil {
			return "", lsifstore.Range{}, false, err
		}
		if exists {
			return text, rn, true, nil
		}
	}

	return "", lsifstore.Range{}, false, nil
}

// Hover returns the hover text and range for the symbol at the given position from the given upload.
func (r *queryResolver) hoverForUpload(ctx context.Context, line, character int, upload dbstore.Dump) (_ string, _ lsifstore.Range, _ bool, err error) {
	// Adjust the path and position for each visible upload based on its git difference to the target commit
	adjustedUpload, ok, err := r.adjustUpload(ctx, line, character, upload)
	if err != nil || !ok {
		return "", lsifstore.Range{}, false, err
	}

	// Fetch hover text from the index
	text, rn, exists, err := r.lsifStore.Hover(
		ctx,
		upload.ID,
		adjustedUpload.AdjustedPathInBundle,
		adjustedUpload.AdjustedPosition.Line,
		adjustedUpload.AdjustedPosition.Character,
	)
	if err != nil || !exists || text == "" {
		return "", lsifstore.Range{}, false, errors.Wrap(err, "lsifStore.Hover")
	}

	// Adjust the highlighted range back to the appropriate range in the target commit
	_, adjustedRange, err := r.adjustRange(ctx, upload.RepositoryID, upload.Commit, r.path, rn)
	if err != nil {
		return "", lsifstore.Range{}, false, err
	}

	return text, adjustedRange, true, nil
}
