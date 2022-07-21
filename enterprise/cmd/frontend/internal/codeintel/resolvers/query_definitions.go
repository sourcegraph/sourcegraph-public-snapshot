package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

const slowDefinitionsRequestThreshold = time.Second

// DefinitionsLimit is maximum the number of locations returned from Definitions.
const DefinitionsLimit = 100

// Definitions returns the list of source locations that define the symbol at the given position.
func (r *queryResolver) Definitions(ctx context.Context, line, character int) (_ []AdjustedLocation, err error) {
	args := shared.RequestArgs{
		RepositoryID: r.repositoryID,
		Commit:       r.commit,
		Path:         r.path,
		Line:         line,
		Character:    character,
	}
	defs, err := r.symbolsResolver.Definitions(ctx, args)
	if err != nil {
		return nil, err
	}

	return uploadLocationToAdjustedLocations(defs), nil
}
