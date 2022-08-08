package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func sharedRangeTolsifstoreRange(r shared.Range) lsifstore.Range {
	return lsifstore.Range{
		Start: lsifstore.Position(r.Start),
		End:   lsifstore.Position(r.End),
	}
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

func sharedDumpToDbstoreUpload(dump shared.Dump) dbstore.Upload {
	return dbstore.Upload{
		ID:                dump.ID,
		Commit:            dump.Commit,
		Root:              dump.Root,
		VisibleAtTip:      dump.VisibleAtTip,
		UploadedAt:        dump.UploadedAt,
		State:             dump.State,
		FailureMessage:    dump.FailureMessage,
		StartedAt:         dump.StartedAt,
		FinishedAt:        dump.FinishedAt,
		ProcessAfter:      dump.ProcessAfter,
		NumResets:         dump.NumResets,
		NumFailures:       dump.NumFailures,
		RepositoryID:      dump.RepositoryID,
		RepositoryName:    dump.RepositoryName,
		Indexer:           dump.Indexer,
		IndexerVersion:    dump.IndexerVersion,
		NumParts:          0,
		UploadedParts:     []int{},
		UploadSize:        nil,
		Rank:              nil,
		AssociatedIndexID: dump.AssociatedIndexID,
	}
}

func sharedRangeToAdjustedRange(rng []shared.AdjustedCodeIntelligenceRange) []AdjustedCodeIntelligenceRange {
	adjustedRange := make([]AdjustedCodeIntelligenceRange, 0, len(rng))
	for _, r := range rng {

		definitions := make([]AdjustedLocation, 0, len(r.Definitions))
		for _, d := range r.Definitions {
			def := AdjustedLocation{
				Dump:           store.Dump(d.Dump),
				Path:           d.Path,
				AdjustedCommit: d.TargetCommit,
				AdjustedRange: lsifstore.Range{
					Start: lsifstore.Position(d.TargetRange.Start),
					End:   lsifstore.Position(d.TargetRange.End),
				},
			}
			definitions = append(definitions, def)
		}

		references := make([]AdjustedLocation, 0, len(r.References))
		for _, d := range r.References {
			ref := AdjustedLocation{
				Dump:           store.Dump(d.Dump),
				Path:           d.Path,
				AdjustedCommit: d.TargetCommit,
				AdjustedRange: lsifstore.Range{
					Start: lsifstore.Position(d.TargetRange.Start),
					End:   lsifstore.Position(d.TargetRange.End),
				},
			}
			references = append(references, ref)
		}

		implementations := make([]AdjustedLocation, 0, len(r.Implementations))
		for _, d := range r.Implementations {
			impl := AdjustedLocation{
				Dump:           store.Dump(d.Dump),
				Path:           d.Path,
				AdjustedCommit: d.TargetCommit,
				AdjustedRange: lsifstore.Range{
					Start: lsifstore.Position(d.TargetRange.Start),
					End:   lsifstore.Position(d.TargetRange.End),
				},
			}
			implementations = append(implementations, impl)
		}

		adj := AdjustedCodeIntelligenceRange{
			Range: lsifstore.Range{
				Start: lsifstore.Position(r.Range.Start),
				End:   lsifstore.Position(r.Range.End),
			},
			Definitions:     definitions,
			References:      references,
			Implementations: implementations,
			HoverText:       r.HoverText,
		}

		adjustedRange = append(adjustedRange, adj)
	}

	return adjustedRange
}

func uploadLocationToAdjustedLocations(location []shared.UploadLocation) []AdjustedLocation {
	uploadLocation := make([]AdjustedLocation, 0, len(location))
	for _, loc := range location {
		dump := store.Dump{
			ID:                loc.Dump.ID,
			Commit:            loc.Dump.Commit,
			Root:              loc.Dump.Root,
			VisibleAtTip:      loc.Dump.VisibleAtTip,
			UploadedAt:        loc.Dump.UploadedAt,
			State:             loc.Dump.State,
			FailureMessage:    loc.Dump.FailureMessage,
			StartedAt:         loc.Dump.StartedAt,
			FinishedAt:        loc.Dump.FinishedAt,
			ProcessAfter:      loc.Dump.ProcessAfter,
			NumResets:         loc.Dump.NumResets,
			NumFailures:       loc.Dump.NumFailures,
			RepositoryID:      loc.Dump.RepositoryID,
			RepositoryName:    loc.Dump.RepositoryName,
			Indexer:           loc.Dump.Indexer,
			IndexerVersion:    loc.Dump.IndexerVersion,
			AssociatedIndexID: loc.Dump.AssociatedIndexID,
		}

		adjustedRange := lsifstore.Range{
			Start: lsifstore.Position{
				Line:      loc.TargetRange.Start.Line,
				Character: loc.TargetRange.Start.Character,
			},
			End: lsifstore.Position{
				Line:      loc.TargetRange.End.Line,
				Character: loc.TargetRange.End.Character,
			},
		}

		uploadLocation = append(uploadLocation, AdjustedLocation{
			Dump:           dump,
			Path:           loc.Path,
			AdjustedCommit: loc.TargetCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return uploadLocation
}
