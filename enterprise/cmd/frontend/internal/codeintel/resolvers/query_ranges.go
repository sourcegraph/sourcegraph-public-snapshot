package resolvers

import (
	"context"
	"time"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

const slowRangesRequestThreshold = time.Second

// Ranges returns code intelligence for the ranges that fall within the given range of lines. These
// results are partial and do not include references outside the current file, or any location that
// requires cross-linking of bundles (cross-repo or cross-root).
func (r *queryResolver) Ranges(ctx context.Context, startLine, endLine int) (adjustedRanges []AdjustedCodeIntelligenceRange, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
	}
	rngs, err := r.symbolsResolver.Ranges(ctx, args, startLine, endLine)
	if err != nil {
		return nil, err
	}

	adjustedRanges = sharedRangeToAdjustedRange(rngs)

	return adjustedRanges, nil
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
