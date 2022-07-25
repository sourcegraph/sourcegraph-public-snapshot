package resolvers

import (
	"context"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

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
