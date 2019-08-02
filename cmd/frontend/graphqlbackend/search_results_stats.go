package graphqlbackend

import (
	"context"
)

func (srs *searchResultsStats) getResults(ctx context.Context) (*searchResultsResolver, error) {
	srs.once.Do(func() {
		srs.srs, srs.srsErr = srs.sr.doResults(ctx, "")
	})
	return srs.srs, srs.srsErr
}
