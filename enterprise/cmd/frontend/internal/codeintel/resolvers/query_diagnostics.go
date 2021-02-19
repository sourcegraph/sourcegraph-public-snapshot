package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const slowDiagnosticsRequestThreshold = time.Second

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *queryResolver) Diagnostics(ctx context.Context, limit int) (adjustedDiagnostics []AdjustedDiagnostic, _ int, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Diagnostics", r.operations.diagnostics, slowDiagnosticsRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("path", r.path),
			log.Int("numUploads", len(r.uploads)),
			log.String("uploads", uploadIDsToString(r.uploads)),
			log.Int("limit", limit),
		},
	})
	defer endObservation()

	adjustedUploads, err := r.adjustUploadPaths(ctx)
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0

	for i := range adjustedUploads {
		traceLog(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		diagnostics, count, err := r.lsifStore.Diagnostics(
			ctx,
			adjustedUploads[i].Upload.ID,
			adjustedUploads[i].AdjustedPathInBundle,
			limit-len(adjustedDiagnostics),
			0,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "lsifStore.Diagnostics")
		}

		for _, diagnostic := range diagnostics {
			adjustedDiagnostic, err := r.adjustDiagnostic(ctx, adjustedUploads[i], diagnostic)
			if err != nil {
				return nil, 0, err
			}

			adjustedDiagnostics = append(adjustedDiagnostics, adjustedDiagnostic)
		}

		totalCount += count
	}

	if len(adjustedDiagnostics) > limit {
		adjustedDiagnostics = adjustedDiagnostics[:limit]
	}
	traceLog(
		log.Int("totalCount", totalCount),
		log.Int("numDiagnostics", len(adjustedDiagnostics)),
	)

	return adjustedDiagnostics, totalCount, nil
}

// adjustUploadPaths adjusts the current target path for each upload visible from the current target
// commit. If an upload cannot be adjusted, it will be omitted from the returned slice.
func (r *queryResolver) adjustUploadPaths(ctx context.Context) ([]adjustedUpload, error) {
	adjustedUploads := make([]adjustedUpload, 0, len(r.uploads))
	for i := range r.uploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, r.uploads[i].Commit, r.path, false)
		if err != nil {
			return nil, errors.Wrap(err, "positionAdjuster.AdjustPath")
		}
		if !ok {
			continue
		}

		adjustedUploads = append(adjustedUploads, adjustedUpload{
			Upload:               r.uploads[i],
			AdjustedPath:         adjustedPath,
			AdjustedPathInBundle: strings.TrimPrefix(adjustedPath, r.uploads[i].Root),
		})
	}

	return adjustedUploads, nil
}

// adjustDiagnostic translates a diagnostic (relative to the indexed commit) into an equivalent diagnostic
// in the requested commit.
func (r *queryResolver) adjustDiagnostic(ctx context.Context, adjustedUpload adjustedUpload, diagnostic lsifstore.Diagnostic) (AdjustedDiagnostic, error) {
	rn := lsifstore.Range{
		Start: lsifstore.Position{
			Line:      diagnostic.StartLine,
			Character: diagnostic.StartCharacter,
		},
		End: lsifstore.Position{
			Line:      diagnostic.EndLine,
			Character: diagnostic.EndCharacter,
		},
	}

	// Adjust path in diagnostic before reading it. This value is used in the adjustRange
	// call below, and is also reflected in the embedded diagnostic value in the return.
	diagnostic.Path = adjustedUpload.Upload.Root + diagnostic.Path

	adjustedCommit, adjustedRange, err := r.adjustRange(
		ctx,
		adjustedUpload.Upload.RepositoryID,
		adjustedUpload.Upload.Commit,
		diagnostic.Path,
		rn,
	)
	if err != nil {
		return AdjustedDiagnostic{}, err
	}

	return AdjustedDiagnostic{
		Diagnostic:     diagnostic,
		Dump:           adjustedUpload.Upload,
		AdjustedCommit: adjustedCommit,
		AdjustedRange:  adjustedRange,
	}, nil
}
