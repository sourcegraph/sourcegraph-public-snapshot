package resolvers

import (
	"context"
	"time"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

const slowImplementationsRequestThreshold = time.Second

// ImplementationsLimit is maximum the number of locations returned from Implementations.
const ImplementationsLimit = 100

// Implementations returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Implementations(ctx context.Context, line, character int, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
		Limit:        limit,
		RawCursor:    rawCursor,
	}
	impl, cursor, err := r.symbolsResolver.Implementations(ctx, args)
	if err != nil {
		return nil, "", err
	}

	adjustedLoc := uploadLocationToAdjustedLocations(impl)

	return adjustedLoc, cursor, nil
}

func symbolDumpToStoreDump(symbolDumps []shared.Dump) []store.Dump {
	dumps := make([]store.Dump, 0, len(symbolDumps))
	for _, d := range symbolDumps {
		dumps = append(dumps, store.Dump{
			ID:                d.ID,
			Commit:            d.Commit,
			Root:              d.Root,
			VisibleAtTip:      d.VisibleAtTip,
			UploadedAt:        d.UploadedAt,
			State:             d.State,
			FailureMessage:    d.FailureMessage,
			StartedAt:         d.StartedAt,
			FinishedAt:        d.FinishedAt,
			ProcessAfter:      d.ProcessAfter,
			NumResets:         d.NumResets,
			NumFailures:       d.NumFailures,
			RepositoryID:      d.RepositoryID,
			RepositoryName:    d.RepositoryName,
			Indexer:           d.Indexer,
			IndexerVersion:    d.IndexerVersion,
			AssociatedIndexID: d.AssociatedIndexID,
		})
	}

	return dumps
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
