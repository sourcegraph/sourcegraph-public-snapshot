package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

// References returns the list of source locations that reference the symbol at the given position.
func (r *queryResolver) References(ctx context.Context, line, character, limit int, rawCursor string) (_ []AdjustedLocation, _ string, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
		Limit:        limit,
		RawCursor:    rawCursor,
	}
	refs, cursor, err := r.symbolsResolver.References(ctx, args)
	if err != nil {
		return nil, "", err
	}

	adjstedLoc := uploadLocationToAdjustedLocations(refs)

	return adjstedLoc, cursor, nil
}
