package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func (srs *searchResultsStats) getResults(ctx context.Context) (*SearchResultsResolver, error) {
	srs.once.Do(func() {
		srs.srs, srs.srsErr = srs.sr.doResults(ctx, result.TypeEmpty)
	})
	return srs.srs, srs.srsErr
}
