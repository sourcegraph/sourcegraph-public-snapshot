package resolvers

import (
	"context"
	"strings"
	"time"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowDiagnosticsRequestThreshold = time.Second

// Diagnostics returns the diagnostics for documents with the given path prefix.
func (r *queryResolver) Diagnostics(ctx context.Context, limit int) (adjustedDiagnostics []AdjustedDiagnostic, _ int, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Limit:        limit,
	}
	diag, cursor, err := r.symbolsResolver.Diagnostics(ctx, args)
	if err != nil {
		return nil, 0, err
	}

	adjustedDiag := sharedDiagnosticAtUploadToAdjustedDiagnostic(diag)

	return adjustedDiag, cursor, nil
}

func sharedDiagnosticAtUploadToAdjustedDiagnostic(shared []shared.DiagnosticAtUpload) []AdjustedDiagnostic {
	adjustedDiagnostics := make([]AdjustedDiagnostic, 0, len(shared))
	for _, diag := range shared {
		diagnosticData := precise.DiagnosticData{
			Severity:       diag.Severity,
			Code:           diag.Code,
			Message:        diag.Message,
			Source:         diag.Source,
			StartLine:      diag.StartLine,
			StartCharacter: diag.StartCharacter,
			EndLine:        diag.EndLine,
			EndCharacter:   diag.EndCharacter,
		}
		lsifDiag := lsifstore.Diagnostic{
			DiagnosticData: diagnosticData,
			DumpID:         diag.DumpID,
			Path:           diag.Path,
		}

		adjusted := AdjustedDiagnostic{
			Diagnostic:     lsifDiag,
			Dump:           store.Dump(diag.Dump),
			AdjustedCommit: diag.AdjustedCommit,
			AdjustedRange: lsifstore.Range{
				Start: lsifstore.Position(diag.AdjustedRange.Start),
				End:   lsifstore.Position(diag.AdjustedRange.End),
			},
		}
		adjustedDiagnostics = append(adjustedDiagnostics, adjusted)
	}
	return adjustedDiagnostics
}

// adjustUploadPaths adjusts the current target path for each upload visible from the current target
// commit. If an upload cannot be adjusted, it will be omitted from the returned slice.
func (r *queryResolver) adjustUploadPaths(ctx context.Context) ([]adjustedUpload, error) {
	adjustedUploads := make([]adjustedUpload, 0, len(r.inMemoryUploads))
	for i := range r.inMemoryUploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, r.inMemoryUploads[i].Commit, r.path, false)
		if err != nil {
			return nil, errors.Wrap(err, "positionAdjuster.AdjustPath")
		}
		if !ok {
			continue
		}

		adjustedUploads = append(adjustedUploads, adjustedUpload{
			Upload:               r.inMemoryUploads[i],
			AdjustedPath:         adjustedPath,
			AdjustedPathInBundle: strings.TrimPrefix(adjustedPath, r.inMemoryUploads[i].Root),
		})
	}

	return adjustedUploads, nil
}
